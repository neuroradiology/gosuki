// TODO: unit test critical error should shutdown the browser
// TODO: shutdown procedure (also close reducer)
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/mozilla"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"git.sp4ke.xyz/sp4ke/gomark/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	QGetBookmarkPlace = `
	SELECT *
	FROM moz_places
	WHERE id = ?
	`
	QBookmarksChanged = `
	SELECT id,type,IFNULL(fk, -1) AS fk,parent,IFNULL(title, '') AS title from moz_bookmarks
	WHERE(lastModified > :last_runtime_utc
		AND lastModified < strftime('%s', 'now')*1000*1000
		AND NOT id IN (:not_root_tags)
		)
	`

	QGetBookmarks = `WITH bookmarks AS
	(SELECT moz_places.url AS url,
			moz_places.description as desc,
			moz_places.title as urlTitle,
			moz_bookmarks.parent AS tagId
		FROM moz_places LEFT OUTER JOIN moz_bookmarks
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent
		IN (SELECT id FROM moz_bookmarks WHERE parent = ? ))

	SELECT url, IFNULL(urlTitle, ''), IFNULL(desc,''),
			tagId, moz_bookmarks.title AS tagTitle

	FROM bookmarks LEFT OUTER JOIN moz_bookmarks
	ON tagId = moz_bookmarks.id
	ORDER BY url`

	// TODO
	QGetBookmarkFolders = `
		SELECT 
		moz_places.id as placesId,
		moz_places.url as url,	
			moz_places.description as description,
			moz_bookmarks.title as title,
			moz_bookmarks.fk ISNULL as isFolder
			
		FROM moz_bookmarks LEFT OUTER JOIN moz_places
		ON moz_places.id = moz_bookmarks.fk
		WHERE moz_bookmarks.parent = 3
	`
)

var ErrInitFirefox = errors.New("Could not start Firefox watcher")

const (
	MozMinJobInterval = 1500 * time.Millisecond
)

type FFBrowser struct {
	BaseBrowser  // embedding
	places       *database.DB
	URLIndexList []string // All elements stored in URLIndex
	tagMap       map[sqlid]*tree.Node
	lastRunTime  time.Time
}

const (
	_ = iota
	BkTypeURL
	BkTypeTagFolder
)

type sqlid int64

const (
	_ = iota
	ffBkRoot
	ffBkMenu
	ffBkToolbar
	ffBkTags
	ffBkOther
	ffBkMobile
)

type AutoIncr struct {
	ID sqlid
}

type FFPlace struct {
	URL         string         `db:"url"`
	Description sql.NullString `db:"description"`
	Title       sql.NullString `db:"title"`
	AutoIncr
}

//type FFBookmark struct {
//BType  int `db:type`
//Parent sqlid
//FK     sql.NullInt64
//Title  sql.NullString
//AutoIncr
//}

type FFBookmark struct {
	btype  sqlid
	parent sqlid
	fk     sqlid
	title  string
	id     sqlid
}

func FFPlacesUpdateHook(op int, db string, table string, rowid int64) {
	fflog.Debug(op)
}

// TODO: Test browser creation errors
// In case of critical errors degrade the browser to only log errors and disable
// all directives
func NewFFBrowser() IBrowser {
	browser := new(FFBrowser)
	browser.Name = "firefox"
	browser.Type = browsers.TFirefox
	browser.BkFile = mozilla.BookmarkFile
	browser.BaseDir = mozilla.GetBookmarkDir()
	browser.UseFileWatcher = true
	browser.Stats = &parsing.Stats{}
	browser.NodeTree = &tree.Node{Name: "root", Parent: nil, Type: "root"}
	browser.tagMap = make(map[sqlid]*tree.Node)

	return browser
}

func (bw *FFBrowser) Shutdown() {
	fflog.Debugf("shutting down ... ")

	if bw.places != nil {

		err := bw.places.Close()
		if err != nil {
			fflog.Critical(err)
		}

	}

	bw.BaseBrowser.Shutdown()
}

func (bw *FFBrowser) Watch() bool {
	if !bw.IsWatching {
		go watch.WatcherThread(bw)
		bw.IsWatching = true
		for _, v := range bw.WatchedPaths {
			fflog.Infof("Watching %s", v)
		}
		return true
	}

	return false
}

func (browser *FFBrowser) copyPlacesToTmp() error {
	err := utils.CopyFilesToTmpFolder(path.Join(browser.BaseDir, browser.BkFile+"*"))
	if err != nil {
		return err
	}

	return nil
}

func (browser *FFBrowser) getPathToPlacesCopy() string {
	return path.Join(utils.TMPDIR, browser.BkFile)
}

func (browser *FFBrowser) InitPlacesCopy() error {
	// Copy places.sqlite to tmp dir
	err := browser.copyPlacesToTmp()
	if err != nil {
		return fmt.Errorf("Could not copy places.sqlite to tmp folder: %s",
			err)
	}

	opts := mozilla.Config.PlacesDSN

	browser.places, err = database.New("places",
		// using the copied places file instead of the original to avoid
		// sqlite vfs lock errors
		browser.getPathToPlacesCopy(),
		database.DBTypeFileDSN, opts).Init()

	if err != nil {
		return err
	}

	return nil
}

func (browser *FFBrowser) Init() error {
	bookmarkPath := path.Join(browser.BaseDir, browser.BkFile)

	// Check if BookmarkPath exists
	exists, err := utils.CheckFileExists(bookmarkPath)
	if err != nil {
		log.Critical(err)
		return ErrInitFirefox
	}

	if !exists {
		return fmt.Errorf("Bookmark path <%s> does not exist", bookmarkPath)
	}

	err = browser.InitPlacesCopy()
	if err != nil {
		return err
	}

	// Setup watcher
	expandedBaseDir, err := filepath.EvalSymlinks(browser.BaseDir)
	if err != nil {
		return err
	}

	browser.WatchedPaths = []string{filepath.Join(expandedBaseDir, "places.sqlite-wal")}

	w := &watch.Watch{
		Path:       expandedBaseDir,
		EventTypes: []fsnotify.Op{fsnotify.Write},
		EventNames: browser.WatchedPaths,
		ResetWatch: false,
	}

	browser.SetupFileWatcher(w)

	/*
	 *Run reducer to avoid duplicate running of jobs
	 *when a batch of events is received
	 */

	browser.InitEventsChan()

	go utils.ReduceEvents(MozMinJobInterval, browser.EventsChan(), browser)

	// Base browser init
	err = browser.BaseBrowser.Init()

	return err
}

func (bw *FFBrowser) Load() error {
	err := bw.BaseBrowser.Load()
	if err != nil {
		return err
	}

	// Parse bookmarks to a flat tree (for compatibility with tree system)
	start := time.Now()
	getFFBookmarks(bw)
	bw.Stats.LastFullTreeParseTime = time.Since(start)
	bw.lastRunTime = time.Now().UTC()

	fflog.Debugf("parsed %d bookmarks and %d nodes in %s",
		bw.Stats.CurrentUrlCount,
		bw.Stats.CurrentNodeCount,
		bw.Stats.LastFullTreeParseTime)
	bw.ResetStats()

	// Sync the URLIndex to the buffer
	// We do not use the NodeTree here as firefox tags are represented
	// as a flat tree which is not efficient, we use the go hashmap instead

	database.SyncURLIndexToBuffer(bw.URLIndexList, bw.URLIndex, bw.BufferDB)

	// Handle empty cache
	if empty, err := CacheDB.IsEmpty(); empty {
		if err != nil {
			fflog.Error(err)
		}
		fflog.Info("cache empty: loading buffer to Cachedb")

		bw.BufferDB.CopyTo(CacheDB)

		fflog.Debugf("syncing <%s> to disk", CacheDB.Name)
	} else {
		bw.BufferDB.SyncTo(CacheDB)
	}
	go CacheDB.SyncToDisk(database.GetDBFullPath())

	// Close the copy places.sqlite
	err = bw.places.Close()

	return err
}

func getFFBookmarks(bw *FFBrowser) {
	//QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"
	//

	rows, err := bw.places.Handle.Query(QGetBookmarks, ffBkTags)
	if err != nil {
		log.Fatal(err)
	}

	// Locked database is critical
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrBusy {
			fflog.Critical(err)
			bw.Shutdown()
			return
		}
	}
	if err != nil {
		fflog.Errorf("%s: %s", bw.places.Name, err)
		return
	}

	// Rebuild node tree
	// Note: the node tree is build only for compatilibity with tree based
	// bookmark parsing.  For efficiency reading after the initial Load() from
	// places.sqlite should be done using a loop instad of tree traversal.
	rootNode := bw.NodeTree

	/*
	 *This pass is used only for fetching bookmarks from firefox.
	 *Checking against the URLIndex should not be done here
	 */
	for rows.Next() {
		var url, title, tagTitle, desc string
		var tagId sqlid
		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		// fflog.Debugf("%s|%s|%s|%d|%s", url, title, desc, tagId, tagTitle)
		if err != nil {
			fflog.Error(err)
		}

		/*
		 * If this is the first time we see this tag
		 * add it to the tagMap and create its node
		 */
		tagNode, tagNodeExists := bw.tagMap[tagId]
		if !tagNodeExists {
			// Add the tag as a node
			tagNode = new(tree.Node)
			tagNode.Type = "tag"
			tagNode.Name = tagTitle
			tagNode.Parent = rootNode
			rootNode.Children = append(rootNode.Children, tagNode)
			bw.tagMap[tagId] = tagNode
			bw.Stats.CurrentNodeCount++
		}

		// Add the url to the tag
		var urlNode *tree.Node
		iUrlNode, urlNodeExists := bw.URLIndex.Get(url)
		if !urlNodeExists {
			urlNode = new(tree.Node)
			urlNode.Type = "url"
			urlNode.URL = url
			urlNode.Name = title
			urlNode.Desc = desc
			bw.URLIndex.Insert(url, urlNode)
			bw.URLIndexList = append(bw.URLIndexList, url)

		} else {
			urlNode = iUrlNode.(*tree.Node)
		}

		// Add tag to urlnode tags
		urlNode.Tags = append(urlNode.Tags, tagNode.Name)

		// Set tag as parent to urlnode
		urlNode.Parent = bw.tagMap[tagId]

		// Add urlnode as child to tag node
		bw.tagMap[tagId].Children = append(bw.tagMap[tagId].Children, urlNode)

		bw.Stats.CurrentUrlCount++
	}
}

func (bw *FFBrowser) fetchUrlChanges(rows *sql.Rows,
	bookmarks map[sqlid]*FFBookmark,
	places map[sqlid]*FFPlace,
) {
	bk := new(FFBookmark)

	// Get the URL that changed
	err := rows.Scan(&bk.id, &bk.btype, &bk.fk, &bk.parent, &bk.title)
	if err != nil {
		log.Fatal(err)
	}

	// database.DebugPrintRow(rows)

	// We found URL change, urls are specified by
	// type == 1
	// fk -> id of url in moz_places
	// parent == tag id
	//
	// Each tag on a url generates 2 or 3 entries in moz_bookmarks
	// 1. If not existing, a (type==2) entry for the tag itself
	// 2. A (type==1) entry for the bookmakred url with (fk -> moz_places.id)
	// 3. A (type==1) (fk-> moz_places.id) (parent == idOf(tag))

	if bk.btype == BkTypeURL {
		var place FFPlace

		// Use unsafe db to ignore non existant columns in
		// dest field
		udb := bw.places.Handle.Unsafe()
		err := udb.QueryRowx(QGetBookmarkPlace, bk.fk).StructScan(&place)
		if err != nil {
			log.Fatal(err)
		}

		fflog.Debugf("Changed URL: %s", place.URL)
		fflog.Debugf("%v", place)

		// put url in the places map
		places[place.ID] = &place
	}

	// This is the tag link
	if bk.btype == BkTypeURL &&
		bk.parent > ffBkMobile {

		bookmarks[bk.id] = bk
	}

	// Tags are specified by:
	// type == 2
	// parent == (Id of root )

	if bk.btype == BkTypeTagFolder {
		bookmarks[bk.id] = bk
	}

	for rows.Next() {
		bw.fetchUrlChanges(rows, bookmarks, places)
	}
	fflog.Debugf("fetching changes done !")
}

func (bw *FFBrowser) Run() {
	startRun := time.Now()

	err := bw.InitPlacesCopy()
	if err != nil {
		fflog.Error(err)
	}

	fflog.Debugf("Checking changes since <%d> %s",
		bw.lastRunTime.UTC().UnixNano()/1000,
		bw.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

	queryArgs := map[string]interface{}{
		"not_root_tags":    []int{ffBkRoot, ffBkTags},
		"last_runtime_utc": bw.lastRunTime.UTC().UnixNano() / 1000,
	}

	query, args, err := sqlx.Named(
		QBookmarksChanged,
		queryArgs,
	)
	if err != nil {
		fflog.Error(err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		fflog.Error(err)
	}

	query = bw.places.Handle.Rebind(query)
	utils.PrettyPrint(query)

	rows, err := bw.places.Handle.Query(query, args...)
	if err != nil {
		fflog.Error(err)
	}

	// Found new results in places db since last time we had changes
	// database.DebugPrintRows(rows) // WARN: This will disable reading rows
	for rows.Next() {
		// next := rows.Next()
		// fflog.Debug("next rows is: ", next)
		// if !next {
		//   break
		// }
		changedURLS := make([]string, 0)

		fflog.Debugf("Found changes since: %s",
			bw.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

		bookmarks := make(map[sqlid]*FFBookmark)
		places := make(map[sqlid]*FFPlace)

		// Fetch all changes into bookmarks and places maps
		bw.fetchUrlChanges(rows, bookmarks, places)

		// utils.PrettyPrint(places)
		// For each url
		for urlId, place := range places {
			var urlNode *tree.Node
			changedURLS = utils.Extends(changedURLS, place.URL)
			iUrlNode, urlNodeExists := bw.URLIndex.Get(place.URL)
			if !urlNodeExists {
				urlNode = new(tree.Node)
				urlNode.Type = "url"
				urlNode.URL = place.URL
				urlNode.Name = place.Title.String
				urlNode.Desc = place.Description.String
				bw.URLIndex.Insert(place.URL, urlNode)

			} else {
				urlNode = iUrlNode.(*tree.Node)
			}

			// First get any new tags
			for bkId, bk := range bookmarks {
				if bk.btype == BkTypeTagFolder &&
					// Ignore root direcotires
					bk.btype != ffBkTags {

					tagNode, tagNodeExists := bw.tagMap[bkId]
					if !tagNodeExists {
						tagNode = new(tree.Node)
						tagNode.Type = "tag"
						tagNode.Name = bk.title
						tagNode.Parent = bw.NodeTree
						bw.NodeTree.Children = append(bw.NodeTree.Children,
							tagNode)
						bw.tagMap[bkId] = tagNode
					}
				}
			}

			// link tags to urls
			for _, bk := range bookmarks {

				// This effectively applies the tag to the URL
				// The tag link should have a parent over 6 and fk->urlId
				fflog.Debugf("Bookmark parent %d", bk.parent)
				if bk.fk == urlId &&
					bk.parent > ffBkMobile {

					// The tag node should have already been created
					tagNode, tagNodeExists := bw.tagMap[bk.parent]
					if tagNodeExists && urlNode != nil {
						fflog.Debugf("URL has tag %s", tagNode.Name)

						urlNode.Tags = utils.Extends(urlNode.Tags, tagNode.Name)

						urlNode.Parent = bw.tagMap[bk.parent]
						tree.Insert(bw.tagMap[bk.parent].Children, urlNode)

						bw.Stats.CurrentUrlCount++
					}
				}
			}

		}

		database.SyncURLIndexToBuffer(changedURLS, bw.URLIndex, bw.BufferDB)
		bw.BufferDB.SyncTo(CacheDB)
		CacheDB.SyncToDisk(database.GetDBFullPath())

	}
	err = rows.Close()
	if err != nil {
		fflog.Error(err)
	}

	// TODO: change logger for more granular debugging

	bw.Stats.LastWatchRunTime = time.Since(startRun)
	// fflog.Debugf("execution time %s", time.Since(startRun))

	go tree.PrintTree(bw.NodeTree) // debugging

	err = bw.places.Close()
	if err != nil {
		fflog.Error(err)
	}

	bw.lastRunTime = time.Now().UTC()
}
