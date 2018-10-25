package main

import (
	"path"
	"time"
)

var Firefox = BrowserPaths{
	"places.sqlite",
	"/home/spike/.mozilla/firefox/p1rrgord.default/",
}

const (
	MozPlacesRootID       = 1
	MozPlacesTagsRootID   = 4
	MozPlacesMobileRootID = 6
)

type FFBrowser struct {
	BaseBrowser //embedding
	_places     *DB
}

type FFTag struct {
	id    int
	title string
}

func NewFFBrowser() IBrowser {
	browser := &FFBrowser{}
	browser.name = "firefox"
	browser.bType = TFirefox
	browser.baseDir = Firefox.BookmarkDir
	browser.bkFile = Firefox.BookmarkFile
	browser.Stats = &ParserStats{}
	browser.NodeTree = &Node{Name: "root", Parent: nil, Type: "root"}

	// Across jobs buffer
	browser.InitBuffer()

	browser.SetupWatcher()

	/*
	 *Run debouncer to avoid duplicate running of jobs
	 *when a batch of events is received
	 */

	//browser.eventsChan = make(chan fsnotify.Event, EventsChanLen)
	//go debouncer(3000*time.Millisecond, browser.eventsChan, browser)

	return browser
}

func (bw *FFBrowser) Watch() bool {

	log.Debugf("<%s> NOT IMPLEMENTED! ", bw.name)
	//if !bw.isWatching {
	//go WatcherThread(bw)
	//bw.isWatching = true
	//return true
	//}

	//return false
	return false
}

func (bw *FFBrowser) Load() {
	bw.BaseBrowser.Load()
	bw.Run()
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

	rows, err := bw._places.Handle.Query(QGetBookmarks, MozPlacesTagsRootID)
	logPanic(err)

	tagMap := make(map[int]*Node)

	// Rebuild node tree
	rootNode := bw.NodeTree

	for rows.Next() {
		var url, title, tagTitle, desc string
		var tagId int
		//tag := &FFTag{}
		//err = rows.Scan(&tag.id, &tag.title)
		err = rows.Scan(&url, &title, &desc, &tagId, &tagTitle)
		logPanic(err)
		//log.Debugf("%s - %s - %s - %d - %s", url, title, desc, tagId, tagTitle)
		//log.Debugf("%s", desc)

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
		urlNode := new(Node)
		urlNode.Type = "url"
		urlNode.URL = url
		urlNode.Name = title
		urlNode.Desc = desc
		urlNode.Parent = tagMap[tagId]
		tagMap[tagId].Children = append(tagMap[tagId].Children, urlNode)

		// Check if url already in index TODO: should be done in new pass
		//iVal, found := bw.URLIndex.Get(urlNode.URL)

		/*
		 * The fields where tags may change are hashed together
		 * to detect changes in futre parses
		 * To handle tag changes we need to get all parent nodes
		 *  (tags) for this url then hash their concatenation
		 */

		//nameHash := xxhash.ChecksumString64(urlNode.Name)
		// TODO: No guarantee we finished gathering tags !!
		// We should check again against index in a new pass
		// This pass needs to finish until we have full count
		// of tags for each bookmark
		//parents := urlNode.GetParentTags()
		//if len(parents) > 4 {
		//tags := make([]string, 0)
		//for _, v := range parents {
		//tags = append(tags, v.Name)
		//}
		////log.Debugf("<%s> --> [%s]", urlNode.URL, strings.Join(tags, "|"))

		//}

		bw.Stats.currentUrlCount++
		bw.Stats.currentNodeCount++
	}

	//go WalkNode(bw.NodeTree)
	//log.Debug(len(tags))

	//bookmarksToBufferFromTags(bw, tag)
}

func bookmarksToBufferFromTag(bw *FFBrowser, tag *FFTag) {
	log.Debugf("db cons: %d", bw._places.Handle.Stats().OpenConnections)

	//log.Debugf("bookmarks for %s", tag.title)
	QGetBookmarksForTag := `SELECT moz_places.url, IFNULL(moz_places.title, '')
						FROM moz_places
						LEFT OUTER JOIN moz_bookmarks
						ON moz_places.id = moz_bookmarks.fk
						WHERE moz_bookmarks.parent = ?`

	rows, err := bw._places.Handle.Query(QGetBookmarksForTag, tag.id)
	logPanic(err)
	//log.Debugf("Query is %s", fmt.Sprintf(QGetBookmarksForTag, tag.id))

	for rows.Next() {
		var url string
		var title string
		err = rows.Scan(&url, &title)
		logPanic(err)
		//log.Debugf("%s ---> %s", tag.title, url)

	}

	//bk := new(Bookmark)
	//bk.Tags = append(bk.Tags, tag.title)
}

func (bw *FFBrowser) Run() {

	log.Debugf("<%s> start bookmark parsing", bw.name)

	// TODO: Node tree is not used for now as the folder
	// parsing is not implemented
	// Rebuild node tree
	// bw.NodeTree = &Node{Name: "root", Parent: nil}

	// Open firefox sqlite db
	bookmarkPath := path.Join(bw.baseDir, bw.bkFile)
	placesDB := DB{}.New("Places", bookmarkPath)
	placesDB.InitRO()
	defer placesDB.Close()

	bw._places = placesDB

	// Parse bookmarks to a flat tree (for compatibility with tree system)
	start := time.Now()
	getFFBookmarks(bw)

	// Finished parsing
	bw.Stats.lastParseTime = time.Since(start)
	log.Debugf("<%s> parsed %d bookmarks and %d nodes", bw.name, bw.Stats.currentUrlCount, bw.Stats.currentNodeCount)
	log.Debugf("<%s> parsed tree in %s", bw.name, bw.Stats.lastParseTime)

	//go PrintTree(bw.NodeTree)

	bw.ResetStats()

}
