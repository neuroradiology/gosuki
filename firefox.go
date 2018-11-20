package main

import (
	"database/sql"
	"fmt"
	"gomark/database"
	"gomark/parsing"
	"gomark/tools"
	"gomark/tree"
	"gomark/watch"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	QGetBookmarkPlace = `
	SELECT id,url,description,title
	FROM moz_places
	WHERE id = ?
	`
	QPlacesDelta = `
	SELECT id,type,IFNULL(fk, -1),parent,IFNULL(title, '') from moz_bookmarks
	WHERE(lastModified > ?
		AND lastModified < strftime('%s', 'now') * 1000 * 1000
		AND NOT id IN (%d,%d)
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
)

var Firefox = BrowserPaths{
	BookmarkFile: "places.sqlite",
	BookmarkDir:  "/home/spike/.mozilla/firefox/p1rrgord.default/",
}

const (
	MozMinJobInterval = 500 * time.Millisecond
)

type FFBrowser struct {
	BaseBrowser  //embedding
	places       *database.DB
	URLIndexList []string // All elements stored in URLIndex
	tagMap       map[int]*tree.Node
	lastRunTime  time.Time
}

const (
	_ = iota
	BkTypeURL
	BkTypeTagFolder
)

type FFBookmarkParent int
type FFBkType int

const (
	_ = iota
	ffBkRoot
	ffBkMenu
	ffBkToolbar
	ffBkTags
	ffBkOther
	ffBkMobile
)

type FFPlace struct {
	id    int
	url   string
	desc  string
	title string
}

type FFBookmark struct {
	id     int
	btype  FFBkType
	parent int
	fk     int
	title  string
}

func FFPlacesUpdateHook(op int, db string, table string, rowid int64) {
	fflog.Debug(op)
}

func NewFFBrowser() IBrowser {
	browser := new(FFBrowser)
	browser.name = "firefox"
	browser.bType = TFirefox
	browser.baseDir = Firefox.BookmarkDir
	browser.bkFile = Firefox.BookmarkFile
	browser.useFileWatcher = true
	browser.Stats = &parsing.Stats{}
	browser.NodeTree = &tree.Node{Name: "root", Parent: nil, Type: "root"}
	browser.tagMap = make(map[int]*tree.Node)

	// Initialize `places.sqlite`
	bookmarkPath := path.Join(browser.baseDir, browser.bkFile)
	browser.places = database.NewRO("Places", bookmarkPath)
	// Buffer that lives accross Run() jobs
	browser.InitBuffer()

	// Setup watcher

	expandedBaseDir, err := filepath.EvalSymlinks(browser.baseDir)

	if err != nil {
		log.Critical(err)
	}

	w := &Watch{
		Path:       expandedBaseDir,
		EventTypes: []fsnotify.Op{fsnotify.Write},
		EventNames: []string{path.Join(expandedBaseDir, "places.sqlite-wal")},
		ResetWatch: false,
	}

	browser.SetupFileWatcher(w)

	/*
	 *Run reducer to avoid duplicate running of jobs
	 *when a batch of events is received
	 */

	browser.eventsChan = make(chan fsnotify.Event, EventsChanLen)

	go tools.ReduceEvents(MozMinJobInterval, browser.eventsChan, browser)

	return browser
}

func (bw *FFBrowser) Shutdown() {

	fflog.Debugf("shutting down ... ")

	err := bw.BaseBrowser.Close()
	if err != nil {
		fflog.Critical(err)
	}

	err = bw.places.Close()
	if err != nil {
		fflog.Critical(err)
	}
}

func (bw *FFBrowser) Watch() bool {

	if !bw.isWatching {
		go watch.WatcherThread(bw)
		bw.isWatching = true
		fflog.Infof("Watching %s", bw.GetPath())
		return true
	}

	return false
}

func (bw *FFBrowser) Load() {
	bw.BaseBrowser.Load()

	// Parse bookmarks to a flat tree (for compatibility with tree system)
	start := time.Now()
	getFFBookmarks(bw)
	bw.Stats.LastFullTreeParseTime = time.Since(start)
	bw.lastRunTime = time.Now().UTC()

	// Finished parsing
	//go PrintTree(bw.NodeTree) // debugging
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
}

func getFFBookmarks(bw *FFBrowser) {

	//QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"
	//

	rows, err := bw.places.Handle.Query(QGetBookmarks, ffBkTags)
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
		var tagId int
		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		//fflog.Debugf("%s|%s|%s|%d|%s", url, title, desc, tagId, tagTitle)
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
	bookmarks map[int]*FFBookmark,
	places map[int]*FFPlace) {

	bk := new(FFBookmark)

	// Get the URL that changed
	rows.Scan(&bk.id, &bk.btype, &bk.fk, &bk.parent, &bk.title)
	fflog.Debug(bk)

	// We found URL change, urls are specified by
	// type == 1
	// fk -> id of url in moz_places
	// parent == tag id
	//
	// Each tag on a url generates 2 or 3 entries in moz_bookmarks
	// 1. If not existing, a (type==2) entry for the tag itself
	// 2. A (type==1) entry for the bookmakred url with (fk -> moz_places.url)
	// 3. A (type==1) (fk-> moz_places.url) (parent == idOf(tag))

	if bk.btype == BkTypeURL {
		place := new(FFPlace)
		res := bw.places.Handle.QueryRow(QGetBookmarkPlace, bk.fk)
		res.Scan(&place.id, &place.url, &place.desc, &place.title)
		fflog.Debugf("Changed URL: %s", place.url)

		// put url in the places map
		places[place.id] = place
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
}

func (bw *FFBrowser) Run() {

	startRun := time.Now()
	fflog.Debugf("Checking changes since %s",
		bw.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

	rows, err := bw.places.Handle.Query(
		// Pre Populate the query
		fmt.Sprintf(QPlacesDelta, "%s", ffBkRoot, ffBkTags),

		// Sql parameter
		bw.lastRunTime.UnixNano()/1000,
	)
	if err != nil {
		fflog.Error(err)
	}
	defer rows.Close()

	// Found new results in places db since last time we had changes
	//database.DebugPrintRows(rows)
	if rows.Next() {
		changedURLS := make([]string, 0)
		bw.lastRunTime = time.Now().UTC()

		//fflog.Debugf("CHANGE ! Time: %s",
		//bw.lastRunTime.Local().Format("Mon Jan 2 15:04:05 MST 2006"))

		bookmarks := make(map[int]*FFBookmark)
		places := make(map[int]*FFPlace)

		// Fetch all changes into bookmarks and places maps
		bw.fetchUrlChanges(rows, bookmarks, places)

		// For each url
		for urlId, place := range places {
			var urlNode *tree.Node
			changedURLS = tools.Extends(changedURLS, place.url)
			iUrlNode, urlNodeExists := bw.URLIndex.Get(place.url)
			if !urlNodeExists {
				urlNode = new(tree.Node)
				urlNode.Type = "url"
				urlNode.URL = place.url
				urlNode.Name = place.title
				urlNode.Desc = place.desc
				bw.URLIndex.Insert(place.url, urlNode)

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
						fflog.Debugf("New tag node %s", tagNode.Name)
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
						//fflog.Debugf("URL has tag %s", tagNode.Name)

						urlNode.Tags = tools.Extends(urlNode.Tags, tagNode.Name)

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

	//TODO: change logger for more granular debugging

	bw.Stats.LastWatchRunTime = time.Since(startRun)
	//fflog.Debugf("execution time %s", time.Since(startRun))
}
