package main

import (
	"io/ioutil"
	"path"
	"time"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/chrome"
	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"git.sp4ke.xyz/sp4ke/gomark/watch"

	"github.com/OneOfOne/xxhash"
	"github.com/buger/jsonparser"
	"github.com/fsnotify/fsnotify"
)

type BaseBrowser = browsers.BaseBrowser
type IBrowser = browsers.IBrowser

//TODO: replace with new profile manager
var ChromeData = browsers.BrowserPaths{
	BookmarkDir: "/home/spike/.config/google-chrome-unstable/Default/",
}

var jsonNodeTypes = struct {
	Folder, URL string
}{"folder", "url"}

var jsonNodePaths = struct {
	Type, Children, URL string
}{"type", "children", "url"}

type ChromeBrowser struct {
	BaseBrowser //embedding
}

type ParseChildJsonFunc func([]byte, jsonparser.ValueType, int, error)
type RecursiveParseJsonFunc func([]byte, []byte, jsonparser.ValueType, int) error

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
			//currentNode.Type = s(value)

		case 1: // name or title
			//currentNode.Name = s(value)
			rawNode.name = value
		case 2:
			//currentNode.URL = s(value)
			rawNode.url = value
		case 3:
			rawNode.children, rawNode.childrenType = value, vt
		}
	}, paths...)
}

// Returns *tree.Node from *RawNode
func (rawNode *RawNode) getNode() *tree.Node {
	node := new(tree.Node)
	node.Type = utils.S(rawNode.nType)
	node.Name = utils.S(rawNode.name)

	return node
}

func NewChromeBrowser() IBrowser {
	browser := new(ChromeBrowser)
	browser.Name = "chrome"
	browser.Type = browsers.TChrome
	browser.BaseDir = ChromeData.BookmarkDir
	browser.BkFile = chrome.BookmarkFile
	browser.Stats = new(parsing.Stats)
	browser.NodeTree = &tree.Node{Name: "root", Parent: nil, Type: "root"}
	browser.UseFileWatcher = true

	// Create watch objects, we will watch the basedir for create events
	watchedEvents := []fsnotify.Op{fsnotify.Create}
	w := &watch.Watch{
		Path:       browser.BaseDir,
		EventTypes: watchedEvents,
		EventNames: []string{path.Join(browser.BaseDir, browser.BkFile)},
		ResetWatch: true,
	}
	browser.SetupFileWatcher(w)

	return browser
}

func (bw *ChromeBrowser) Watch() bool {
	if !bw.IsWatching {
		go watch.WatcherThread(bw)
		bw.IsWatching = true
		log.Infof("<%s> Watching %s", bw.Name, bw.GetBookmarksPath())
		return true
	}

	return false
}

func (bw *ChromeBrowser) Init() error {
	return bw.BaseBrowser.Init()
}

func (bw *ChromeBrowser) Load() error {

	// BaseBrowser load method
	err := bw.BaseBrowser.Load()
	if err != nil {
		return err
	}

	bw.Run()

	return nil
}

func (bw *ChromeBrowser) Run() {
	startRun := time.Now()

	// Rebuild node tree
	bw.RebuildNodeTree()

	// Load bookmark file
	bookmarkPath := path.Join(bw.BaseDir, bw.BkFile)
	f, err := ioutil.ReadFile(bookmarkPath)
	if err != nil {
		log.Critical(err)
	}

	var parseChildren ParseChildJsonFunc
	var jsonParseRecursive RecursiveParseJsonFunc

	parseChildren = func(childVal []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			log.Panic(err)
		}

		jsonParseRecursive(nil, childVal, dataType, offset)
	}

	// Needed to store the parent of each child node
	var parentNodes []*tree.Node

	jsonParseRoots := func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		bw.Stats.CurrentNodeCount++
		rawNode := new(RawNode)
		rawNode.parseItems(node)
		//log.Debugf("Parsing root folder %s", rawNode.name)

		currentNode := rawNode.getNode()

		// Process this node as parent node later
		parentNodes = append(parentNodes, currentNode)

		// add the root node as parent to this node
		currentNode.Parent = bw.NodeTree

		// Add this root node as a child of the root node
		bw.NodeTree.Children = append(bw.NodeTree.Children, currentNode)

		// Call recursive parsing of this node which must
		// a root folder node
		jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

		// Finished parsing this root, it is not anymore a parent
		_, parentNodes = parentNodes[len(parentNodes)-1], parentNodes[:len(parentNodes)-1]

		//log.Debugf("Parsed root %s folder", rawNode.name)

		return nil
	}

	// Main recursive parsing function that parses underneath
	// each root folder
	jsonParseRecursive = func(key []byte, node []byte, dataType jsonparser.ValueType, offset int) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		bw.Stats.CurrentNodeCount++

		rawNode := new(RawNode)
		rawNode.parseItems(node)

		currentNode := rawNode.getNode()
		//log.Debugf("parsing node %s", currentNode.Name)

		// if parents array is not empty
		if len(parentNodes) != 0 {
			parent := parentNodes[len(parentNodes)-1]
			//log.Debugf("Adding current node to parent %s", parent.Name)

			// Add current node to closest parent
			currentNode.Parent = parent

			// Add current node as child to parent
			currentNode.Parent.Children = append(currentNode.Parent.Children, currentNode)
		}

		// if node is a folder with children
		if rawNode.childrenType == jsonparser.Array && len(rawNode.children) > 2 { // if len(children) > len("[]")

			//log.Debugf("Started folder %s", rawNode.name)
			parentNodes = append(parentNodes, currentNode)

			// Process recursively all child nodes of this folder node
			jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

			//log.Debugf("Finished folder %s", rawNode.name)
			_, parentNodes = parentNodes[len(parentNodes)-1], parentNodes[:len(parentNodes)-1]

		}

		// if node is url(leaf), handle the url
		if utils.S(rawNode.nType) == jsonNodeTypes.URL {

			currentNode.URL = utils.S(rawNode.url)
			bw.Stats.CurrentUrlCount++
			// Check if url-node already in index
			var nodeVal *tree.Node
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
				//log.Debugf("URL Found in index")
				nodeVal = iVal.(*tree.Node)

				// hash(name) is different meaning new commands/tags could
				// be added, we need to process the parsing hoos
				if nodeVal.NameHash != nameHash {
					//log.Debugf("URL name changed !")

					// Run parse hooks on node
					bw.RunParseHooks(currentNode)
				}

				// Else we do nothing, the node will not
				// change
			}

			//If parent is folder, add it as tag and add current node as child
			//And add this link as child
			if currentNode.Parent.Type == jsonNodeTypes.Folder {
				//log.Debug("Parent is folder, parsing as tag ...")
				currentNode.Tags = append(currentNode.Tags, currentNode.Parent.Name)
			}

		}

		return nil
	}

	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	// Start a new node tree building job
	jsonparser.ObjectEach(rootsData, jsonParseRoots)
	bw.Stats.LastFullTreeParseTime = time.Since(startRun)
	log.Debugf("<%s> parsed tree in %s", bw.Name, bw.Stats.LastFullTreeParseTime)
	// Finished node tree building job

	// Debug walk tree
	//go PrintTree(bw.NodeTree)

	// Reset the index to represent the nodetree
	bw.RebuildIndex()

	// Finished parsing
	log.Debugf("<%s> parsed %d bookmarks and %d nodes", bw.Name, bw.Stats.CurrentUrlCount, bw.Stats.CurrentNodeCount)
	// Reset parser counter
	bw.ResetStats()

	//Add nodeTree to Cache
	//log.Debugf("<%s> buffer content", bw.Name)
	//bw.BufferDB.Print()

	log.Debugf("<%s> syncing to buffer", bw.Name)
	database.SyncTreeToBuffer(bw.NodeTree, bw.BufferDB)
	log.Debugf("<%s> tree synced to buffer", bw.Name)

	//bw.BufferDB.Print()

	// cacheDB represents bookmarks across all browsers
	// From browsers it should support: add/update
	// Delete method should only be possible through admin interface
	// We could have an @ignore command to ignore a bookmark

	// URLIndex is a hashmap index of all URLS representing current state
	// of the browser

	// nodeTree is current state of the browser as tree

	// Buffer is the current state of the browser represetned by
	// URLIndex and nodeTree

	// If cacheDB is empty just copy buffer to cacheDB
	// until local db is already populated and preloaded
	//debugPrint("%d", BufferDB.Count())
	if empty, err := CacheDB.IsEmpty(); empty {
		if err != nil {
			log.Error(err)
		}
		log.Info("cache empty: loading buffer to Cachedb")

		bw.BufferDB.CopyTo(CacheDB)

		log.Debugf("syncing <%s> to disk", CacheDB.Name)
	} else {
		bw.BufferDB.SyncTo(CacheDB)
	}

	go CacheDB.SyncToDisk(database.GetDBFullPath())
	bw.Stats.LastWatchRunTime = time.Since(startRun)
}
