package main

import (
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"time"

	"github.com/buger/jsonparser"
)

const (
	RE_TAGS = `\B#\w+`
)

type Bookmark struct {
	url      string
	metadata string
	tags     []string
	desc     string
	modifed  time.Time
	//flags int
}

type parserStat struct {
	lastNodeCount    int
	lastUrlCount     int
	currentNodeCount int
	currentUrlCount  int
}

var jsonNodeTypes = struct {
	Folder, Url string
}{"folder", "url"}

var jsonNodePaths = struct {
	Type, Children, Url string
}{"type", "children", "url"}

type parseFunc func([]byte, []byte, jsonparser.ValueType, int) error

func _s(value interface{}) string {
	return string(value.([]byte))
}

func findTagsInTitle(title []byte) {
	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAll(title, -1)
	debugPrint("%s ---> found following tags: %s", title, tags)
}

func googleParseBookmarks(bw *bookmarkWatcher) {

	// Create buffer db
	//bufferDB := DB{"buffer", DB_BUFFER_PATH, nil, false}
	bufferDB := DB{}.New("buffer", DB_BUFFER_PATH)
	defer bufferDB.Close()
	bufferDB.Init()

	// Load bookmark file
	bookmarkPath := path.Join(bw.baseDir, bw.bkFile)
	f, err := ioutil.ReadFile(bookmarkPath)
	logPanic(err)

	var parseChildren func([]byte, jsonparser.ValueType, int, error)
	var gJsonParseRecursive func([]byte, []byte, jsonparser.ValueType, int) error

	parseChildren = func(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Panic(err)
		}

		gJsonParseRecursive(nil, childVal, dataType, offset)
	}

	gJsonParseRecursive = func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {
		// Core of google chrome bookmark parsing
		// Any loading to local db is done here
		bw.stats.currentNodeCount++

		var nodeType, children []byte
		var childrenType jsonparser.ValueType
		bookmark := &Bookmark{}

		// Paths to lookup in node payload
		paths := [][]string{
			[]string{"type"},
			[]string{"name"}, // Title of page
			[]string{"url"},
			[]string{"children"},
		}

		jsonparser.EachKey(node, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
			switch idx {
			case 0:
				nodeType = value
			case 1: // name or title
				bookmark.metadata = _s(value)
			case 2:
				bookmark.url = _s(value)
			case 3:
				children, childrenType = value, vt
			}
		}, paths...)

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		// if node is url(leaf), handle the url
		if _s(nodeType) == jsonNodeTypes.Url {
			// Add bookmark to db here
			//debugPrint("%s", url)
			//debugPrint("%s", node)

			// Find tags in title
			//findTagsInTitle(name)
			bw.stats.currentUrlCount++
			addBookmark(bookmark, bufferDB)

		}

		// if node is a folder with children
		if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
			jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)
		}

		return nil
	}

	//debugPrint("parsing bookmarks")
	// Begin parsing
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	debugPrint("loading bookmarks to bufferdb")
	// Load bookmarks to currentJobDB
	jsonparser.ObjectEach(rootsData, gJsonParseRecursive)

	// Finished parsing
	debugPrint("parsed %d bookmarks", bw.stats.currentUrlCount)

	// Reset parser counter
	bw.stats.lastUrlCount = bw.stats.currentUrlCount
	bw.stats.lastNodeCount = bw.stats.currentNodeCount
	bw.stats.currentNodeCount = 0
	bw.stats.currentUrlCount = 0

	// Compare currentDb with memCacheDb for new bookmarks

	// If CACHE_DB is empty just copy bufferDB to CACHE_DB
	// until local db is already populated and preloaded
	//debugPrint("%d", bufferDB.Count())
	if empty, err := CACHE_DB.isEmpty(); empty {
		logPanic(err)
		//debugPrint("first preloading, copying bufferdb to cachedb")

		//start := time.Now()
		bufferDB.SyncTo(CACHE_DB)
		//debugPrint("<%s> is now (%d)", CACHE_DB.name, CACHE_DB.Count())
		//elapsed := time.Since(start)
		//debugPrint("copy in %s", elapsed)

		debugPrint("syncing <%s> to disk", CACHE_DB.name)
		CACHE_DB.SyncToDisk(getDBFullPath())
	}

	//_ = CACHE_DB.Print()
}

func addBookmark(bookmark *Bookmark, db *DB) {
	// TODO
	// Single out unique urls
	//debugPrint("%v", bookmark)
	_db := db.handle

	tx, err := _db.Begin()
	logPanic(err)

	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	logPanic(err)
	defer stmt.Close()

	_, err = stmt.Exec(bookmark.url, bookmark.metadata, "", "", 0)
	logPanic(err)

	err = tx.Commit()
	logPanic(err)

}
