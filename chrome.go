package main

import (
	"io/ioutil"
	"path"

	"github.com/buger/jsonparser"
)

var jsonNodeTypes = struct {
	Folder, URL string
}{"folder", "url"}

var jsonNodePaths = struct {
	Type, Children, URL string
}{"type", "children", "url"}

type ParseChildFunc func([]byte, jsonparser.ValueType, int, error)
type RecursiveParseFunc func([]byte, []byte, jsonparser.ValueType, int) error

type ChromeBrowser struct {
	BaseBrowser //embedding
}

func NewChromeBrowser() IBrowser {
	browser := &ChromeBrowser{}
	browser.name = "chrome"
	browser.bType = TChrome
	browser.baseDir = Chrome.BookmarkDir
	browser.bkFile = Chrome.BookmarkFile
	browser.stats = &ParserStats{}

	browser.SetupWatcher()

	return browser
}

func (bw *ChromeBrowser) Watch() bool {
	if !bw.isWatching {
		go WatcherThread(bw)
		bw.isWatching = true
		return true
	}

	return false
}

func (bw *ChromeBrowser) Load() {

	// Check if cache is initialized
	if cacheDB == nil || cacheDB.handle == nil {
		log.Critical("cache is not yet initialized !")
		panic("cache is not yet initialized !")
	}

	if bw.watcher == nil {
		log.Fatal("watcher not initialized SetupWatcher() !")
	}

	log.Debug("preloading bookmarks")
	bw.Run()
}

func (bw *ChromeBrowser) Run() {

	// Create buffer db
	//bufferDB := DB{"buffer", DB_BUFFER_PATH, nil, false}
	bw.InitBuffer()
	defer bw.bufferDB.Close()

	// Load bookmark file
	bookmarkPath := path.Join(bw.baseDir, bw.bkFile)
	f, err := ioutil.ReadFile(bookmarkPath)
	logPanic(err)

	var parseChildren ParseChildFunc
	var gJsonParseRecursive RecursiveParseFunc

	parseChildren = func(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Panic(err)
		}

		gJsonParseRecursive(nil, childVal, dataType, offset)
	}

	rootsNode := new(Node)
	currentNode := rootsNode

	gJsonParseRecursive = func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {
		// Core of google chrome bookmark parsing
		// Any loading to local db is done here
		bw.stats.currentNodeCount++

		parentNode := currentNode
		currentNode := new(Node)
		currentNode.Parent = parentNode

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
				currentNode.Type = _s(value)

			case 1: // name or title
				currentNode.Name = _s(value)
			case 2:
				currentNode.URL = _s(value)
			case 3:
				children, childrenType = value, vt
			}
		}, paths...)

		bookmark.Metadata = currentNode.Name
		bookmark.URL = currentNode.URL

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		// if node is url(leaf), handle the url
		if _s(nodeType) == jsonNodeTypes.URL {
			// Add bookmark to db here
			//debugPrint("%s", url)
			//debugPrint("%s", node)

			// Find tags in title
			//findTagsInTitle(name)
			bw.stats.currentUrlCount++

			// Run parsehoos before adding bookmark
			bw.RunParseHooks(bookmark)

			// Add bookmark
			bookmark.add(bw.bufferDB)
		}

		parentNode.Children = append(parentNode.Children, currentNode)

		// if node is a folder with children
		if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
			jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)
		}

		return nil
	}

	//debugPrint("parsing bookmarks")
	// Begin parsing
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	log.Debug("loading bookmarks to bufferdb")
	// Load bookmarks to currentJobDB
	jsonparser.ObjectEach(rootsData, gJsonParseRecursive)

	// Debug walk tree
	//go WalkNode(rootsNode)

	// Finished parsing
	log.Debugf("parsed %d bookmarks", bw.stats.currentUrlCount)

	// Reset parser counter
	bw.stats.lastURLCount = bw.stats.currentUrlCount
	bw.stats.lastNodeCount = bw.stats.currentNodeCount
	bw.stats.currentNodeCount = 0
	bw.stats.currentUrlCount = 0

	// Compare currentDb with memCacheDb for new bookmarks

	// If cacheDB is empty just copy bufferDB to cacheDB
	// until local db is already populated and preloaded
	//debugPrint("%d", bufferDB.Count())
	if empty, err := cacheDB.isEmpty(); empty {
		logPanic(err)
		log.Debug("cache empty: loading bufferdb to cachedb")

		//start := time.Now()
		bw.bufferDB.SyncTo(cacheDB)
		//debugPrint("<%s> is now (%d)", cacheDB.name, cacheDB.Count())
		//elapsed := time.Since(start)
		//debugPrint("copy in %s", elapsed)

		debugPrint("syncing <%s> to disk", cacheDB.name)
		cacheDB.SyncToDisk(getDBFullPath())
	}

	// TODO: Check if new/modified bookmarks in buffer compared to cache
	log.Debugf("TODO: check if new/modified bookmarks in %s compared to %s", bw.bufferDB.name, cacheDB.name)

}
