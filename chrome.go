package main

import (
	"io/ioutil"
	"path"

	"github.com/OneOfOne/xxhash"
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
	browser.nodeTree = &Node{Name: "root", Parent: nil}
	browser.cNode = browser.nodeTree

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

	bw.InitIndex()

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

	gJsonParseRecursive = func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {
		// Core of google chrome bookmark parsing
		// Any loading to local db is done here
		bw.stats.currentNodeCount++

		//log.Debugf("moving current node %v as parent", currentNode.Name)
		currentNode := new(Node)

		currentNode.Parent = bw.cNode
		bw.cNode.Children = append(bw.cNode.Children, currentNode)
		bw.cNode = currentNode

		var nodeType, nodeName, nodeURL, children []byte
		var childrenType jsonparser.ValueType

		//log.Debugf("parent %v", parentNode)

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
				//currentNode.Type = _s(value)

			case 1: // name or title
				//currentNode.Name = _s(value)
				nodeName = value
			case 2:
				//currentNode.URL = _s(value)
				nodeURL = value
			case 3:
				children, childrenType = value, vt
			}
		}, paths...)

		log.Debugf("parsing node %s", nodeName)

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		// if node is a folder with children
		if childrenType == jsonparser.Array && len(children) > 2 { // if len(children) > len("[]")
			jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

			// Finished parsing all children
			// Add them into current node ?
		}

		currentNode.Type = _s(nodeType)
		currentNode.Name = _s(nodeName)

		// if node is url(leaf), handle the url
		if _s(nodeType) == jsonNodeTypes.URL {

			currentNode.URL = _s(nodeURL)

			bw.stats.currentUrlCount++

			// Check if url-node already in index
			var nodeVal *Node
			iVal, found := bw.URLIndex.Get(currentNode.URL)

			nameHash := xxhash.ChecksumString64(currentNode.Name)
			// If node url not in index, add it to index
			if !found {
				//log.Debugf("Not found")

				// store hash(name)
				currentNode.NameHash = nameHash

				// The value in the index will be a
				// pointer to currentNode
				//log.Debugf("Inserting url %s to index", nodeURL)
				bw.URLIndex.Insert(currentNode.URL, currentNode)

				// If we find the node already in index
				// we check if the hash(name) changed  meaning
				// the data changed
			} else {
				//log.Debugf("Found")
				nodeVal = iVal.(*Node)

				// hash(name) is different, we will update the
				// index and parse the bookmark
				if nodeVal.NameHash != nameHash {

					// Update node in index
					currentNode.NameHash = nameHash

					if currentNode.NameHash != nodeVal.NameHash {
						panic("currentNode.NameHash != indexValue.NameHash")
					}

					// Run parse hooks on node
					bw.RunParseHooks(currentNode)

				}

				// Else we do nothing, the node will not
				// change
			}

			// If parent is folder, add it as tag and add current node as child
			// And add this link as child
			if currentNode.Parent.Type == jsonNodeTypes.Folder {
				log.Debug("Parent is folder, parsing as tag ...")
				currentNode.Tags = append(currentNode.Tags, currentNode.Parent.Name)
			}

		}

		//log.Debugf("Adding current node %v to parent %v", currentNode.Name, parentNode)
		//parentNode.Children = append(parentNode.Children, currentNode)
		//currentNode.Parent = parentNode

		return nil
	}

	//debugPrint("parsing bookmarks")
	// Begin parsing
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	log.Debug("loading bookmarks to index")

	jsonparser.ObjectEach(rootsData, gJsonParseRecursive)

	// Debug walk tree
	go WalkNode(bw.nodeTree)

	// Finished parsing
	log.Debugf("parsed %d bookmarks", bw.stats.currentUrlCount)

	// Reset parser counter
	bw.stats.lastURLCount = bw.stats.currentUrlCount
	bw.stats.lastNodeCount = bw.stats.currentNodeCount
	bw.stats.currentNodeCount = 0
	bw.stats.currentUrlCount = 0

	// Compare currentDb with index for new bookmarks

	log.Debug("TODO: Compare cacheDB with index")

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
