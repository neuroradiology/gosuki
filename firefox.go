package main

import (
	"fmt"
	"path"
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
	browser.NodeTree = &Node{Name: "root", Parent: nil}

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

func getTags(bw *FFBrowser) {
	var tags []*FFTag
	QGetTags := "SELECT id,title from moz_bookmarks WHERE parent = %d"

	rows, err := bw._places.Handle.Query(fmt.Sprintf(QGetTags, MozPlacesTagsRootID))
	logPanic(err)

	for rows.Next() {
		tag := &FFTag{}
		err = rows.Scan(&tag.id, &tag.title)
		logPanic(err)
		tags = append(tags, tag)
	}
	log.Debug(len(tags))
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

	// WITH bookmarks AS (SELECT moz_places.url , moz_bookmarks.parent AS tagId FROM moz_places LEFT OUTER JOIN moz_bookmarks ON moz_places.id = moz_bookmarks.fk WHERE moz_bookmarks.parent IN (SELECT id FROM moz_bookmarks WHERE parent = 4 )) SELECT url, tagId, moz_bookmarks.title FROM bookmarks LEFT OUTER JOIN moz_bookmarks ON tagId = moz_bookmarks.id ORDER BY moz_bookmarks.title WHERE title = 'bitcoin'
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

	// Start parsing from the root node (id = 1, type = 2) and down the tree

	// First get all tags and register them as nodes under the root node
	getTags(bw)
	log.Debugf("<%s> finished parsing tags", bw.name)

}
