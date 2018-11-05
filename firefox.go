package main

import (
	"database/sql"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
)

var Firefox = BrowserPaths{
	BookmarkFile: "places.sqlite",
	BookmarkDir:  "/home/spike/.mozilla/firefox/p1rrgord.default/",
}

const (
	MozPlacesRootID       = 1
	MozPlacesTagsRootID   = 4
	MozPlacesMobileRootID = 6
	MozMinJobInterval     = 1 * time.Second
)

type FFBrowser struct {
	BaseBrowser //embedding
	places      *DB
	// TODO: Use URLIndex instead
	URLIndexList []string  // All elements stored in URLIndex
	qChanges     *sql.Stmt // Last changes query
	lastRunTime  time.Time
}

type FFTag struct {
	id    int
	title string
}

func FFPlacesUpdateHook(op int, db string, table string, rowid int64) {
	log.Debug(op)
}

func NewFFBrowser() IBrowser {
	browser := &FFBrowser{}
	browser.name = "firefox"
	browser.bType = TFirefox
	browser.baseDir = Firefox.BookmarkDir
	browser.bkFile = Firefox.BookmarkFile
	browser.useFileWatcher = true
	browser.Stats = &ParserStats{}
	browser.NodeTree = &Node{Name: "root", Parent: nil, Type: "root"}

	// Initialize `places.sqlite`
	bookmarkPath := path.Join(browser.baseDir, browser.bkFile)
	browser.places = DB{}.New("Places", bookmarkPath)
	browser.places.InitRO()

	// Buffer that lives accross Run() jobs
	browser.InitBuffer()

	// Setup watcher

	w := &Watch{
		path:       path.Join(browser.baseDir),
		eventTypes: []fsnotify.Op{fsnotify.Write},
		eventNames: []string{path.Join(browser.baseDir, "places.sqlite-wal")},
		resetWatch: false,
	}

	browser.SetupFileWatcher(w)

	/*
	 *Run reducer to avoid duplicate running of jobs
	 *when a batch of events is received
	 */

	browser.eventsChan = make(chan fsnotify.Event, EventsChanLen)

	go reducer(MozMinJobInterval, browser.eventsChan, browser)

	// Prepare sql statements
	// Check changed urls in DB
	// Firefox records time UTC and microseconds
	// Sqlite converts automatically from utc to local
	QPlacesDelta := `
	SELECT * from moz_bookmarks
	WHERE(lastModified > ?
		AND lastModified < strftime('%s', 'now') * 1000 * 1000)
	`
	stmt, err := browser.places.handle.Prepare(QPlacesDelta)
	if err != nil {
		log.Error(err)
	}
	browser.qChanges = stmt

	return browser
}

func (bw *FFBrowser) Shutdown() {

	log.Debugf("<%s> shutting down ... ", bw.name)

	err := bw.qChanges.Close()
	if err != nil {
		log.Critical(err)
	}

	err = bw.BaseBrowser.Close()
	if err != nil {
		log.Critical(err)
	}

	err = bw.places.Close()
	if err != nil {
		log.Critical(err)
	}
}

func (bw *FFBrowser) Watch() bool {

	log.Debugf("<%s> TODO ... ", bw.name)

	if !bw.isWatching {
		go WatcherThread(bw)
		bw.isWatching = true
		log.Infof("<%s> Watching %s", bw.name, bw.GetPath())
		return true
	}

	return false
}

func (bw *FFBrowser) Load() {
	bw.BaseBrowser.Load()

	// Parse bookmarks to a flat tree (for compatibility with tree system)
	start := time.Now()
	getFFBookmarks(bw)
	bw.Stats.lastParseTime = time.Since(start)
	bw.lastRunTime = time.Now().UTC()

	// Finished parsing
	//go PrintTree(bw.NodeTree) // debugging
	log.Debugf("<%s> parsed %d bookmarks and %d nodes in %s", bw.name,
		bw.Stats.currentUrlCount, bw.Stats.currentNodeCount, bw.Stats.lastParseTime)

	bw.ResetStats()

	// Sync the URLIndex to the buffer
	// We do not use the NodeTree here as firefox tags are represented
	// as a flat tree which is not efficient, we use the go hashmap instead

	syncURLIndexToBuffer(bw.URLIndexList, bw.URLIndex, bw.BufferDB)

	bw.BufferDB.SyncTo(CacheDB)
}

func getFFBookmarks(bw *FFBrowser) {

	QGetBookmarks := `WITH bookmarks AS

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

	//QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"

	rows, err := bw.places.handle.Query(QGetBookmarks, MozPlacesTagsRootID)
	if err != nil {
		log.Error(err)
	}

	tagMap := make(map[int]*Node)

	// Rebuild node tree
	// Note: the node tree is build only for compatilibity
	// pruposes with tree based bookmark parsing.
	// For efficiency reading after the initial parse from places.sqlite,
	// reading should be done in loops in instead of tree parsing.
	rootNode := bw.NodeTree

	/*
	 *This pass is used only for fetching bookmarks from firefox.
	 *Checking against the URLIndex should not be done here
	 */
	for rows.Next() {
		var url, title, tagTitle, desc string
		var tagId int
		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		//log.Debugf("%s|%s|%s|%d|%s", url, title, desc, tagId, tagTitle)
		if err != nil {
			log.Error(err)
		}

		/*
		 * If this is the first time we see this tag
		 * add it to the tagMap and create its node
		 */
		tagNode, tagNodeExists := tagMap[tagId]
		if !tagNodeExists {
			// Add the tag as a node
			tagNode = new(Node)
			tagNode.Type = "tag"
			tagNode.Name = tagTitle
			tagNode.Parent = rootNode
			rootNode.Children = append(rootNode.Children, tagNode)
			tagMap[tagId] = tagNode
			bw.Stats.currentNodeCount++
		}

		// Add the url to the tag
		var urlNode *Node
		iUrlNode, urlNodeExists := bw.URLIndex.Get(url)
		if !urlNodeExists {
			urlNode = new(Node)
			urlNode.Type = "url"
			urlNode.URL = url
			urlNode.Name = title
			urlNode.Desc = desc
			bw.URLIndex.Insert(url, urlNode)
			bw.URLIndexList = append(bw.URLIndexList, url)

		} else {
			urlNode = iUrlNode.(*Node)
		}

		// Add tag to urlnode tags
		urlNode.Tags = append(urlNode.Tags, tagNode.Name)

		// Set tag as parent to urlnode
		urlNode.Parent = tagMap[tagId]

		// Add urlnode as child to tag node
		tagMap[tagId].Children = append(tagMap[tagId].Children, urlNode)

		bw.Stats.currentUrlCount++
		bw.Stats.currentNodeCount++
	}

	/*
	 *Build tags for each url then check against URLIndex
	 *for changes
	 */

	// Check if url already in index TODO: should be done in new pass
	//iVal, found := bw.URLIndex.Get(urlNode.URL)

	/*
	 * The fields where tags may change are hashed together
	 * to detect changes in futre parses
	 * To handle tag changes we need to get all parent nodes
	 *  (tags) for this url then hash their concatenation
	 */

	//nameHash := xxhash.ChecksumString64(urlNode.Name)

}

func (bw *FFBrowser) Run() {

	//log.Debugf("%d", bw.lastRunTime.Unix())
	//var _time string
	//row := bw.places.handle.QueryRow("SELECT strftime('%s', 'now') AS now")
	//row.Scan(&_time)
	//log.Debug(_time)

	log.Debugf("Checking changes since %s",
		bw.lastRunTime.Format("Mon Jan 2 15:04:05 MST 2006"))
	start := time.Now()
	rows, err := bw.qChanges.Query(bw.lastRunTime.UnixNano() / 1000)
	if err != nil {
		log.Error(err)
	}
	defer rows.Close()
	elapsed := time.Since(start)
	log.Debugf("Places test query in %s", elapsed)

	// Found new results in places db since last time we had changes
	if rows.Next() {
		bw.lastRunTime = time.Now().UTC()
		log.Debugf("<%s> CHANGE ! Time: %s", bw.name,
			bw.lastRunTime.Format("Mon Jan 2 15:04:05 MST 2006"))
		//DebugPrintRows(rows)
	} else {
		log.Debugf("<%s> no change", bw.name)
	}

	//TODO: change logger for more granular debugging
	// candidates:  glg

}
