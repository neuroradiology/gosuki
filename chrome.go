package main

import (
	"io/ioutil"
	"path"
	"time"

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

type RawNode struct {
	name         []byte
	nType        []byte
	url          []byte
	children     []byte
	childrenType jsonparser.ValueType
}

func (rawNode *RawNode) parseItems(nodeData []byte) {

	// Paths to lookup in node payload
	paths := [][]string{
		[]string{"type"},
		[]string{"name"}, // Title of page
		[]string{"url"},
		[]string{"children"},
	}

	jsonparser.EachKey(nodeData, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			rawNode.nType = value
			//currentNode.Type = _s(value)

		case 1: // name or title
			//currentNode.Name = _s(value)
			rawNode.name = value
		case 2:
			//currentNode.URL = _s(value)
			rawNode.url = value
		case 3:
			rawNode.children, rawNode.childrenType = value, vt
		}
	}, paths...)
}

// Returns *Node from *RawNode
func (rawNode *RawNode) getNode() *Node {
	node := new(Node)
	node.Type = _s(rawNode.nType)
	node.Name = _s(rawNode.name)

	return node
}

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

	// Across jobs buffer
	browser.InitBuffer()

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

	// BaseBrowser load method
	bw.BaseBrowser.Load()

	bw.Run()
}

func (bw *ChromeBrowser) Run() {

	// Rebuild node tree
	bw.nodeTree = &Node{Name: "root", Parent: nil}

	// Load bookmark file
	bookmarkPath := path.Join(bw.baseDir, bw.bkFile)
	f, err := ioutil.ReadFile(bookmarkPath)
	logPanic(err)

	var parseChildren ParseChildFunc
	var jsonParseRecursive RecursiveParseFunc

	parseChildren = func(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Panic(err)
		}

		jsonParseRecursive(nil, childVal, dataType, offset)
	}

	// Needed to store the parent of each child node
	var parentNodes []*Node

	jsonParseRoots := func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		bw.stats.currentNodeCount++

		rawNode := new(RawNode)
		rawNode.parseItems(node)
		log.Debugf("Parsing root folder %s", rawNode.name)

		currentNode := rawNode.getNode()

		// Process this node as parent node later
		parentNodes = append(parentNodes, currentNode)

		// add the root node as parent to this node
		currentNode.Parent = bw.nodeTree

		// Add this root node as a child of the root node
		bw.nodeTree.Children = append(bw.nodeTree.Children, currentNode)

		// Call recursive parsing of this node which must
		// a root folder node
		jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

		// Finished parsing this root, it is not anymore a parent
		_, parentNodes = parentNodes[len(parentNodes)-1], parentNodes[:len(parentNodes)-1]

		log.Debugf("Parsed root %s folder", rawNode.name)

		return nil
	}

	// Main recursive parsing function that parses underneath
	// each root folder
	jsonParseRecursive = func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		bw.stats.currentNodeCount++

		rawNode := new(RawNode)
		rawNode.parseItems(node)

		currentNode := rawNode.getNode()
		log.Debugf("parsing node %s", currentNode.Name)

		// if parents array is not empty
		if len(parentNodes) != 0 {
			parent := parentNodes[len(parentNodes)-1]
			log.Debugf("Adding current node to parent %s", parent.Name)

			// Add current node to closest parent
			currentNode.Parent = parent

			// Add current node as child to parent
			currentNode.Parent.Children = append(currentNode.Parent.Children, currentNode)
		}

		// if node is a folder with children
		if rawNode.childrenType == jsonparser.Array && len(rawNode.children) > 2 { // if len(children) > len("[]")

			log.Debugf("Started folder %s", rawNode.name)
			parentNodes = append(parentNodes, currentNode)

			// Process recursively all child nodes of this folder node
			jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

			log.Debugf("Finished folder %s", rawNode.name)
			_, parentNodes = parentNodes[len(parentNodes)-1], parentNodes[:len(parentNodes)-1]

		}

		// if node is url(leaf), handle the url
		if _s(rawNode.nType) == jsonNodeTypes.URL {

			currentNode.URL = _s(rawNode.url)

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

				// Run tag parsing hooks
				bw.RunParseHooks(currentNode)

				// If we find the node already in index
				// we check if the hash(name) changed  meaning
				// the data changed
			} else {
				log.Debugf("URL Found in index")
				nodeVal = iVal.(*Node)

				// hash(name) is different meaning new commands/tags could
				// be added, we need to mark this bookmark as `has_changed`
				if nodeVal.NameHash != nameHash {
					//log.Debugf("URL name changed !")

					// Mark current node (BK) as changed
					currentNode.HasChanged = true

					// Run parse hooks on node
					bw.RunParseHooks(currentNode)

				}

				// Else we do nothing, the node will not
				// change
			}

			//If parent is folder, add it as tag and add current node as child
			//And add this link as child
			if currentNode.Parent.Type == jsonNodeTypes.Folder {
				log.Debug("Parent is folder, parsing as tag ...")
				currentNode.Tags = append(currentNode.Tags, currentNode.Parent.Name)
			}

		}

		return nil
	}

	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	// Start a new node tree building job
	start := time.Now()
	jsonparser.ObjectEach(rootsData, jsonParseRoots)
	elapsed := time.Since(start)
	log.Debugf("Parsing tree in %s", elapsed)
	// Finished node tree building job

	// Debug walk tree
	//go WalkNode(bw.nodeTree)

	// Reset the index to represent the nodetree
	bw.RebuildIndex()

	// Finished parsing
	log.Debugf("parsed %d bookmarks", bw.stats.currentUrlCount)

	// Reset parser counter
	bw.stats.lastURLCount = bw.stats.currentUrlCount
	bw.stats.lastNodeCount = bw.stats.currentNodeCount
	bw.stats.currentNodeCount = 0
	bw.stats.currentUrlCount = 0

	//Add nodeTree to Cache
	log.Debugf("Buffer content")
	bw.bufferDB.Print()

	log.Debugf("syncing to buffer")
	syncTreeToBuffer(bw.nodeTree, bw.bufferDB)
	log.Debugf("Tree synced to buffer")

	bw.bufferDB.Print()

	// cacheDB represents bookmarks across all browsers
	// From browsers it should support: add/update
	// Delete method should only be possible through admin interface
	// We could have an @ignore command to ignore a bookmark

	// URLIndex is a hashmap index of all URLS representing current state
	// of the browser

	// nodeTree is current state of the browser as tree

	// Buffer is the current state of the browser represetned by
	// URLIndex and nodeTree

	log.Debug("TODO: Compare cacheDB with index")

	// If cacheDB is empty just copy buffer to cacheDB
	// until local db is already populated and preloaded
	//debugPrint("%d", bufferDB.Count())
	if empty, err := cacheDB.isEmpty(); empty {
		logPanic(err)
		log.Debug("cache empty: loading buffer to cachedb")

		//start := time.Now()
		bw.bufferDB.SyncTo(cacheDB)
		//debugPrint("<%s> is now (%d)", cacheDB.name, cacheDB.Count())
		//elapsed := time.Since(start)
		//debugPrint("copy in %s", elapsed)

		debugPrint("syncing <%s> to disk", cacheDB.name)
		cacheDB.SyncToDisk(getDBFullPath())
	}

	// Implement incremental sync by doing INSERTs
	bw.bufferDB.SyncTo(cacheDB)

	cacheDB.SyncToDisk(getDBFullPath())

	// TODO: Check if new/modified bookmarks in buffer compared to cache
	log.Debugf("TODO: check if new/modified bookmarks in %s compared to %s", bw.bufferDB.name, cacheDB.name)

}
