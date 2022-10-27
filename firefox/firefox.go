// TODO: unit test critical error should shutdown the browser
// TODO: shutdown procedure (also close reducer)
// TODO: migrate own commands to here
// TODO: handle flag management from this package

// TODO: Refactoring:
// TODO: * Implement Init() and Load() for firefox
package firefox

import (
	"database/sql"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"git.sp4ke.xyz/sp4ke/gomark/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// sql queries
const (
	QGetBookmarkPlace = `
	SELECT *
	FROM moz_places
	WHERE id = ?
	`
	//TEST:
	QBookmarksChanged = `
    SELECT id,type,IFNULL(fk, -1) AS fk,parent,IFNULL(title, '') AS title from moz_bookmarks
    WHERE(lastModified > :last_runtime_utc
        AND lastModified < strftime('%s', 'now')*1000*1000
        AND NOT id IN (:not_root_tags)
    )
	`

	//TEST:
	QgetBookmarks = `
    WITH bookmarks AS
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
	ORDER BY url
    `

	//TEST:
	//TODO:
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

var (
	ErrInitFirefox = errors.New("could not start Firefox watcher")
	log            = logging.GetLogger("FF")
)

const (
	MozMinJobInterval = 1500 * time.Millisecond
)

// moz_bookmarks.type
const (
	_ = iota
	BkTypeURL
	BkTypeTagFolder
)

type sqlid int64

// moz_bookmarks.id
const (
	_           = iota // 0
	ffBkRoot           // 1
	ffBkMenu           // 2 Main bookmarks menu
	ffBkToolbar        // 3 Bk tookbar that can be toggled under URL zone
	ffBkTags           // 4 Hidden menu used for tags, stored as a flat one level menu
	ffBkOther          // 5 Most bookmarks are automatically stored here
	ffBkMobile         // 6 Mobile bookmarks stored here by default
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

type Firefox struct {
	*FirefoxConfig

	// sqlite con to places.sqlite
	places *database.DB

	// All elements stored in URLIndex
	URLIndexList []string

	// Map from place tag IDs to the parse node tree
	tagMap map[sqlid]*tree.Node

	lastRunTime time.Time
}

// FIX: depends on config which should be initialized before this init
func init() {
	browsers.RegisterBrowser(Firefox{FirefoxConfig: FFConfig})
	//TIP: cmd.RegisterModCommand(BrowserName, &cli.Command{
	//TIP: 	Name: "test",
	//TIP: })
	//TIP: cmd.RegisterModCommand(BrowserName, &cli.Command{
	//TIP: 	Name: "test2",
	//TIP: })
}

func (ff Firefox) ModInfo() browsers.ModInfo {
	return browsers.ModInfo{
		ID:  browsers.ModID(ff.Name),
		New: func() browsers.Module { return new(Firefox) },
	}
}

// TEST: Test browser creation errors
// In case of critical errors degrade the browser to only log errors and disable
// all directives
// func NewFirefox() *Firefox {
// 	browser := &FFBrowser{
// 		BaseBrowser: browsers.BaseBrowser{
// 			Name:   BrowserName,
// 			Type:   browsers.TFirefox,
// 			BkFile: mozilla.BookmarkFile,
// 			//FIX: basedir is set here before config has time to set the right path
// 			//NOTE: maybe use reflect and instanciate the module in core package
// 			//              &bookmarkDir ??
// 			BaseDir:        getBookmarkDir(),
// 			NodeTree:       &tree.Node{Name: "root", Parent: nil, Type: "root"},
// 			Stats:          &parsing.Stats{},
// 			UseFileWatcher: true,
// 		},
// 		tagMap: make(map[sqlid]*tree.Node),
// 	}
//
// 	return browser
// }

// TEST:
// Implements browser.Initializer interface
func (f *Firefox) Init() error {
	log.Infof("initializing <%s>", f.Name)
	bookmarkPath, err := f.BookmarkPath()
	if err != nil {
		return err
	}

	log.Debugf("bookmark path is: %s", bookmarkPath)

	err = f.initPlacesCopy()
	if err != nil {
		return err
	}

	// Setup watcher
	expandedBaseDir, err := filepath.EvalSymlinks(f.BkDir)
	if err != nil {
		return err
	}

	w := &watch.Watch{
		Path:       expandedBaseDir,
		EventTypes: []fsnotify.Op{fsnotify.Write},
		EventNames: []string{filepath.Join(expandedBaseDir, "places.sqlite-wal")},
		ResetWatch: false,
	}

	browsers.SetupWatchersWithReducer(f.BrowserConfig, browsers.ReducerChanLen, w)

	/*
	 *Run reducer to avoid duplicate jobs when a batch of events is received
	 */
	// TODO!: make a new copy of places for every new event change

	// Add a reducer to the watcher
	go watch.ReduceEvents(MozMinJobInterval, f)

	return nil
}

func (f *Firefox) Watcher() *watch.WatchDescriptor {
	return f.Watcher()
}

// Firefox custom logic for preloading the bookmarks when the browser module
// starts. Implements browsers.Loader interface.
// TODO!: Implement default loading in browsers package
func (f *Firefox) Load() error {
	// err := f.BaseBrowser.Load()

	// Parse bookmarks to a flat tree (for compatibility with tree system)
	start := time.Now()
	getBookmarks(f)
	f.Stats.LastFullTreeParseTime = time.Since(start)
	f.lastRunTime = time.Now().UTC()

	log.Debugf("parsed %d bookmarks and %d nodes in %s",
		f.Stats.CurrentUrlCount,
		f.Stats.CurrentNodeCount,
		f.Stats.LastFullTreeParseTime)
	f.Stats.Reset()

	// Sync the URLIndex to the buffer
	// We do not use the NodeTree here as firefox tags are represented
	// as a flat tree which is not efficient, we use the go hashmap instead

	database.SyncURLIndexToBuffer(f.URLIndexList, f.URLIndex, f.BufferDB)

	// Handle empty cache
	if empty, err := database.Cache.DB.IsEmpty(); empty {
		if err != nil {
			return err
		}
		log.Info("cache empty: loading buffer to Cachedb")

		f.BufferDB.CopyTo(database.Cache.DB)

		log.Debugf("syncing <%s> to disk", database.Cache.DB.Name)
	} else {
		f.BufferDB.SyncTo(database.Cache.DB)
	}

	database.Cache.DB.SyncToDisk(database.GetDBFullPath())

	//DEBUG:
	tree.PrintTree(f.NodeTree)

	// Close the copy places.sqlite
	err := f.places.Close()

	return err
}

// Implement browsers.Runner interface
func (f *Firefox) Run() {
	startRun := time.Now()

	err := f.initPlacesCopy()
	if err != nil {
		log.Error(err)
	}

	log.Debugf("Checking changes since <%d> %s",
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
		log.Error(err)
	}

	query, args, err = sqlx.In(query, args...)
	if err != nil {
		log.Error(err)
	}

	query = bw.places.Handle.Rebind(query)
	utils.PrettyPrint(query)

	rows, err := bw.places.Handle.Query(query, args...)
	if err != nil {
		log.Error(err)
	}

	// Found new results in places db since last time we had changes
	// database.DebugPrintRows(rows) // WARN: This will disable reading rows
	// TEST: implement this in a func and unit test it
	// NOTE: this looks like a lot of code reuse in fetchUrlChanges()
	for rows.Next() {
		// next := rows.Next()
		// log.Debug("next rows is: ", next)
		// if !next {
		//   break
		// }
		changedURLS := make([]string, 0)

		log.Debugf("Found changes since: %s",
			bw.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

		// extract bookmarks to this map
		bookmarks := make(map[sqlid]*FFBookmark)

		// record new places to this map
		places := make(map[sqlid]*FFPlace)

		// Fetch all changes into bookmarks and places maps
		bw.fetchUrlChanges(rows, bookmarks, places)

		// utils.PrettyPrint(places)
		// For each url
		for urlId, place := range places {
			var urlNode *tree.Node
			changedURLS = utils.Extends(changedURLS, place.URL)

			ok, urlNode := bw.addUrlNode(place.URL, place.Title.String, place.Description.String)
			if !ok {
				log.Infof("url <%s> already in url index", place.URL)
			}

			// First get any new bookmarks
			for bkId, bk := range bookmarks {

				// if bookmark type is folder or tag
				if bk.btype == BkTypeTagFolder &&

					// Ignore root directories
					// NOTE: ffBkTags change time shows last time bookmark tags
					// whre changed ?
					bkId != ffBkTags {

					log.Debugf("adding tag node %s", bk.title)
					ok, tagNode := bw.addTagNode(bkId, bk.title)
					if !ok {
						log.Infof("tag <%s> already in tag map", tagNode.Name)
					}
				}
			}

			// link tags(moz_bookmark) to urls (moz_places)
			for _, bk := range bookmarks {

				// This effectively applies the tag to the URL
				// The tag link should have a parent over 6 and fk->urlId
				log.Debugf("Bookmark parent %d", bk.parent)
				if bk.fk == urlId &&
					bk.parent > ffBkMobile {

					// The tag node should have already been created
					tagNode, tagNodeExists := bw.tagMap[bk.parent]

					if tagNodeExists && urlNode != nil {
						log.Debugf("URL has tag %s", tagNode.Name)

						urlNode.Tags = utils.Extends(urlNode.Tags, tagNode.Name)

						tree.AddChild(bw.tagMap[bk.parent], urlNode)
						//TEST: remove after testing this code section
						// urlNode.Parent = bw.tagMap[bk.parent]
						// tree.Insert(bw.tagMap[bk.parent].Children, urlNode)

						bw.Stats.CurrentUrlCount++
					}
				}
			}

		}

		database.SyncURLIndexToBuffer(changedURLS, bw.URLIndex, bw.BufferDB)
		bw.BufferDB.SyncTo(database.Cache.DB)
		database.Cache.DB.SyncToDisk(database.GetDBFullPath())

	}
	err = rows.Close()
	if err != nil {
		log.Error(err)
	}

	// TODO: change logger for more granular debugging

	bw.Stats.LastWatchRunTime = time.Since(startRun)
	// log.Debugf("execution time %s", time.Since(startRun))

	// tree.PrintTree(bw.NodeTree) // debugging

	err = bw.places.Close()
	if err != nil {
		log.Error(err)
	}

	bw.lastRunTime = time.Now().UTC()
}

// Implement browsers.Shutdowner
func (f *Firefox) Shutdown() {
	log.Debugf("shutting down ... ")

	if f.places != nil {

		err := f.places.Close()
		if err != nil {
			log.Critical(err)
		}
	}
}

func (browser *Firefox) copyPlacesToTmp() error {
	err := utils.CopyFilesToTmpFolder(path.Join(f.config.BkDir, browser.BkFile+"*"))
	if err != nil {
		return err
	}

	return nil
}

func (browser *Firefox) getPathToPlacesCopy() string {
	return path.Join(utils.TMPDIR, browser.BkFile)
}

// TEST:
// HACK: addUrl and addTag share a lot of code, find a way to reuse shared code
// and only pass extra details about tag/url along in some data structure
// PROBLEM: tag nodes use IDs and URL nodes use URL as hashes
func (browser *Firefox) addUrlNode(url, title, desc string) (bool, *tree.Node) {
	var urlNode *tree.Node
	iUrlNode, exists := browser.URLIndex.Get(url)
	if !exists {
		urlNode := &tree.Node{
			Name: title,
			Type: "url",
			URL:  url,
			Desc: desc,
		}

		log.Debugf("inserting url %s in url index", url)
		browser.URLIndex.Insert(url, urlNode)
		browser.URLIndexList = append(browser.URLIndexList, url)

		return true, urlNode
	} else {
		urlNode = iUrlNode.(*tree.Node)
	}

	return false, urlNode
}

// adds a new tagNode if it is not existing in the tagMap
// returns true if tag added or false if already existing
// returns the created tagNode
func (browser *Firefox) addTagNode(tagId sqlid, tagName string) (bool, *tree.Node) {
	// node, exists :=
	node, exists := browser.tagMap[tagId]
	if exists {
		return false, node
	}

	tagNode := &tree.Node{
		Name:   tagName,
		Type:   "tag",
		Parent: browser.NodeTree, // root node
	}

	tree.AddChild(browser.NodeTree, tagNode)
	browser.tagMap[tagId] = tagNode
	browser.Stats.CurrentNodeCount++

	return true, tagNode
}

// Copies places.sqlite to a tmp dir to read a VFS lock sqlite db
func (f *Firefox) initPlacesCopy() error {
	err := f.copyPlacesToTmp()
	if err != nil {
		return fmt.Errorf("Could not copy places.sqlite to tmp folder: %s",
			err)
	}

	opts := FFConfig.PlacesDSN

	f.places, err = database.New("places",
		// using the copied places file instead of the original to avoid
		// sqlite vfs lock errors
		f.getPathToPlacesCopy(),
		database.DBTypeFileDSN, opts).Init()

	if err != nil {
		return err
	}

	return nil
}

// load all bookmarks from `places.sqlite` and store them in BaseBrowser.NodeTree
// this method is used the first time gomark is started or to extract bookmarks
// using a command
func getBookmarks(bw *Firefox) {
	log.Debugf("root tree children len is %d", len(bw.NodeTree.Children))
	//QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"
	//

	rows, err := bw.places.Handle.Query(QgetBookmarks, ffBkTags)
	if err != nil {
		log.Fatal(err)
	}

	// Locked database is critical
	if e, ok := err.(sqlite3.Error); ok {
		if e.Code == sqlite3.ErrBusy {
			log.Critical(err)
			bw.Shutdown()
			return
		}
	}
	if err != nil {
		log.Errorf("%s: %s", bw.places.Name, err)
		return
	}

	// Rebuilding node tree
	// Note: the node tree is built only for compatilibity with tree based
	// bookmark parsing and might later be useful for debug/UI features.
	// For efficiency reading after the initial Load() from
	// places.sqlite should be done using a loop instad of tree traversal.

	/*
	 *This pass is used only for fetching bookmarks from firefox.
	 *Checking against the URLIndex should not be done here
	 */
	for rows.Next() {
		var url, title, tagTitle, desc string
		var tagId sqlid

		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		// log.Debugf("%s|%s|%s|%d|%s", url, title, desc, tagId, tagTitle)
		if err != nil {
			log.Error(err)
		}

		/*
		 * If this is the first time we see this tag
		 * add it to the tagMap and create its node
		 */
		ok, tagNode := bw.addTagNode(tagId, tagTitle)
		if !ok {
			log.Infof("tag <%s> already in tag map", tagNode.Name)
		}

		// Add the url to the tag
		ok, urlNode := bw.addUrlNode(url, title, desc)
		if !ok {
			log.Infof("url <%s> already in url index", url)
		}

		// Add tag name to urlnode tags
		urlNode.Tags = append(urlNode.Tags, tagNode.Name)

		// Set tag as parent to urlnode
		tree.AddChild(bw.tagMap[tagId], urlNode)

		bw.Stats.CurrentUrlCount++
	}

	log.Debugf("root tree children len is %d", len(bw.NodeTree.Children))
}

// fetchUrlChanges method  
// scan rows from a firefox `places.sqlite` db and extract all bookmarks and
// places (moz_bookmarks, moz_places tables) that changed/are new since the browser.lastRunTime
// using the QBookmarksChanged query
func (bw *Firefox) fetchUrlChanges(rows *sql.Rows,
	bookmarks map[sqlid]*FFBookmark,
	places map[sqlid]*FFPlace,
) {
	bk := &FFBookmark{}

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

		log.Debugf("Changed URL: %s", place.URL)
		log.Debugf("%v", place)

		// put url in the places map
		places[place.ID] = &place
	}

	// This is the tag link
	if bk.btype == BkTypeURL &&
		// ignore original tags/folder from mozilla
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
	log.Debugf("fetching changes done !")
}
