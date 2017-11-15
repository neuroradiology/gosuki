package main

import (
	"io/ioutil"
	"log"
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
	//flags int
}

var parserStat = struct {
	lastNodeCount int
	lastUrlCount  int
	currentCount  int
}{}

var jsonNodeTypes = struct {
	Folder, Url string
}{"folder", "url"}

var jsonNodePaths = struct {
	Type, Children, Url string
}{"type", "children", "url"}

type parseFunc func([]byte, []byte, jsonparser.ValueType, int) error

func parseChildren(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
	if err != nil {
		log.Panic(err)
	}

	gJsonParseRecursive(nil, childVal, dataType, offset)
}

func _s(value interface{}) string {
	return string(value.([]byte))
}

func findTagsInTitle(title []byte) {
	var regex = regexp.MustCompile(RE_TAGS)
	tags := regex.FindAll(title, -1)
	debugPrint("%s ---> found following tags: %s", title, tags)
}

func googleParseBookmarks(bookmarkPath string) {
	// Load bookmark file
	f, err := ioutil.ReadFile(BOOKMARK_FILE)

	if err != nil {
		log.Fatal(err)
	}

	//debugPrint("parsing bookmarks")
	// Begin parsing
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	debugPrint("loading bookmarks to currentJobDB")

	// Load bookmarks to currentJobDB
	jsonparser.ObjectEach(rootsData, gJsonParseRecursive)

	// Finished parsing
	debugPrint("parsed %d bookmarks", parserStat.lastUrlCount)

	// Compare currentDb with memCacheDb for new bookmarks

	// If memCacheDb is empty just copy data to memCacheDb
	if isEmptyDb(memCacheDb) {
		debugPrint("first preloading, copying jobdb to cachedb")

		start := time.Now()
		printDBCount(memCacheDb)
		syncToDB(DB_CURRENT, DB_MEMCACHE)
		printDBCount(memCacheDb)
		elapsed := time.Since(start)
		debugPrint("%s", elapsed)
	}

}

func gJsonParseRecursive(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {
	// Core of google chrome bookmark parsing
	// Any loading to local db is done here
	parserStat.lastNodeCount++

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
		parserStat.lastUrlCount++
		addBookmark(bookmark)

	}

	// if node is a folder with children
	if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
		jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)
	}

	return nil
}

func addBookmark(bookmark *Bookmark) {
	// TODO
	// Single out unique urls
	//debugPrint("%v", bookmark)

	tx, err := currentJobDB.Begin()
	logPanic(err)

	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	logPanic(err)
	defer stmt.Close()

	_, err = stmt.Exec(bookmark.url, bookmark.metadata, "", "", 0)
	logPanic(err)

	err = tx.Commit()
	logPanic(err)

}
