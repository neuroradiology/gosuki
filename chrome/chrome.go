package chrome

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/OneOfOne/xxhash"
	"github.com/fsnotify/fsnotify"

	"github.com/buger/jsonparser"

	"git.blob42.xyz/gomark/gosuki/database"
	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/modules"
	"git.blob42.xyz/gomark/gosuki/tree"
	"git.blob42.xyz/gomark/gosuki/watch"
)

var (
	log = logging.GetLogger("Chrome")
)

type ParseChildJSONFunc func([]byte, jsonparser.ValueType, int, error)
type RecursiveParseJSONFunc func([]byte, []byte, jsonparser.ValueType, int) error

var jsonNodeTypes = map[string]tree.NodeType{
	"folder": tree.FolderNode,
	"url":    tree.URLNode,
}

var jsonNodePaths = struct {
	Type, Children, URL string
}{"type", "children", "url"}

// stores json nodes in memory
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
		{"type"},
		{"name"}, // Title of page
		{"url"},
		{"children"},
	}

	jsonparser.EachKey(nodeData, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		if err != nil {
			log.Critical("error parsing node items")
		}

		switch idx {
		case 0:
			rawNode.nType = value

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
	nType, ok := jsonNodeTypes[string(rawNode.nType)]
	if !ok {
		log.Criticalf("unknown node type: %s", rawNode.nType)
	}
	node.Type = nType

	node.Name = string(rawNode.name)

	return node
}

// Chrome browser module
type Chrome struct {
	// holds browsers.BrowserConfig
	*ChromeConfig
}

// Init() is the first method called after a browser instance is created
// and registered.
// Return ok, error
func (ch *Chrome) Init(_ *modules.Context) error {
	log.Infof("initializing <%s>", ch.Name)
	return ch.setupWatchers()
}

func (ch *Chrome) setupWatchers() error {
	bookmarkDir, err := ch.BookmarkDir()
	log.Debugf("Watching path: %s", bookmarkDir)
	if err != nil {
		return err
	}
	bookmarkPath := filepath.Join(bookmarkDir, ch.BkFile)
	// Setup watcher
	w := &watch.Watch{
		Path:       bookmarkDir,
		EventTypes: []fsnotify.Op{fsnotify.Create},
		EventNames: []string{bookmarkPath},


		// NOTE: it used to be that chrome watcher would go stale after the
		// first event, this is because the bookmark file is created after the
		// browser is started, so we need to reset the watcher after the first
		// event. This is not needed anymore as we are using the watcher to
		// watch the bookmark dir and not the file itself.
		ResetWatch: false,
	}

	ok, err := modules.SetupWatchers(ch.BrowserConfig, w)
	if err != nil {
		return fmt.Errorf("could not setup watcher: %s", err)
	}
	if !ok {
		return errors.New("could not setup watcher")
	}

	return nil
}

func (ch *Chrome) ResetWatcher() error {
	w := ch.GetWatcher()
	if err := w.W.Close(); err != nil {
		return err
	}
	if err := ch.setupWatchers(); err != nil {
		return err
	}

	go watch.WatcherThread(ch)
	return nil
}

// Returns a pointer to an initialized browser config
func (ch Chrome) Config() *modules.BrowserConfig {
	return ch.BrowserConfig
}

func (ch Chrome) ModInfo() modules.ModInfo {
	return modules.ModInfo{
		ID: modules.ModID(ch.Name),
		New: func() modules.Module {
			return NewChrome()
		},
	}
}

func (ch *Chrome) Watch() *watch.WatchDescriptor {
	// calls modules.BrowserConfig.GetWatcher()
	return ch.GetWatcher()
}

func (ch *Chrome) Run() {
	startRun := time.Now()

	// Rebuild node tree
	ch.NodeTree = &tree.Node{
		Name:   RootNodeName,
		Parent: nil,
		Type:   tree.RootNode,
	}

	// Load bookmark file
	//WIP:  use builting path helpers
	bookmarkPath, err := ch.BookmarkPath()
	if err != nil {
		log.Critical(err)
		return
	}

	f, err := os.ReadFile(bookmarkPath)
	if err != nil {
		log.Critical(err)
		return
	}

	var parseChildren ParseChildJSONFunc
	var jsonParseRecursive RecursiveParseJSONFunc

	parseChildren = func(childVal []byte,
		dataType jsonparser.ValueType,
		offset int,
		err error) {
		if err != nil {
			log.Panic(err)
		}

		err = jsonParseRecursive(nil, childVal, dataType, offset)
		if err != nil {
			log.Critical(err)
		}

	}

	// Needed to store the parent of each child node
	var parentNodes []*tree.Node

	jsonParseRoots := func(key []byte,
		node []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		ch.CurrentNodeCount++
		rawNode := new(RawNode)
		rawNode.parseItems(node)

		//log.Debugf("Parsing root folder %s", rawNode.name)

		currentNode := rawNode.getNode()

		// Process this node as parent node later
		parentNodes = append(parentNodes, currentNode)

		// add the root node as parent to this node
		currentNode.Parent = ch.NodeTree

		// Add this root node as a child of the root node
		ch.NodeTree.Children = append(ch.NodeTree.Children, currentNode)

		// Call recursive parsing of this node which must
		// a root folder node
		jsonparser.ArrayEach(node, parseChildren, jsonNodePaths.Children)

		// Finished parsing this root, it is not anymore a parent
		_, parentNodes = parentNodes[len(parentNodes)-1],
			parentNodes[:len(parentNodes)-1]

		//log.Debugf("Parsed root %s folder", rawNode.name)

		return nil
	}

	// Main recursive parsing underneath each root folder
	jsonParseRecursive = func(key []byte,
		node []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {

		// If node type is string ignore (needed for sync_transaction_version)
		if dataType == jsonparser.String {
			return nil
		}

		ch.CurrentNodeCount++

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
		if currentNode.Type == tree.URLNode {

			currentNode.URL = string(rawNode.url)
			ch.CurrentURLCount++

			// Check if url-node already in index
			var nodeVal *tree.Node
			iVal, found := ch.URLIndex.Get(currentNode.URL)

			nameHash := xxhash.ChecksumString64(currentNode.Name)

			// If node url not in index, add it to index
			if !found {
				//log.Debugf("Not found")

				// store hash(name)
				currentNode.NameHash = nameHash

				// The value in the index will be a
				// pointer to currentNode
				log.Debugf("Inserting url %s to index", currentNode.URL)
				ch.URLIndex.Insert(currentNode.URL, currentNode)

				// Run registered bookmark parsing hooks
				err = ch.CallHooks(currentNode)
				if err != nil {
					return err
				}

				// If we find the node already in index
				// we check if the hash(name) changed  meaning
				// the data changed
			} else {
				// log.Debugf("URL Found in index")
				nodeVal = iVal.(*tree.Node)

				// hash(name) is different meaning new commands/tags could
				// be added, we need to process the parsing hoos
				if nodeVal.NameHash != nameHash {
					log.Debugf("URL name changed !")

					// Run parse hooks on node
					ch.CallHooks(currentNode)
				}

				// Else we do nothing, the node will not
				// change
			}

			//If parent is folder, add it as tag and add current node as child
			//And add this link as child
			if currentNode.Parent.Type == tree.FolderNode {
				log.Debug("Parent is folder, parsing as tag ...")
				currentNode.Tags = append(currentNode.Tags, currentNode.Parent.Name)
			}
		}

		return nil
	}

	// starts from the "roots" key of chrome json bookmark file
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	// Start a new node tree building job
	jsonparser.ObjectEach(rootsData, jsonParseRoots)
	ch.LastFullTreeParseTime = time.Since(startRun)
	log.Debugf("<%s> parsed tree in %s", ch.Name, ch.LastFullTreeParseTime)
	// Finished node tree building job

	// Debug walk tree
	//go PrintTree(ch.NodeTree)

	// Reset the index to represent the nodetree
	ch.RebuildIndex()

	// Finished parsing
	log.Debugf("<%s> parsed %d bookmarks and %d nodes", ch.Name, ch.CurrentURLCount, ch.CurrentNodeCount)

	//Add nodeTree to Cache
	//log.Debugf("<%s> buffer content", ch.Name)
	//ch.BufferDB.Print()

	log.Debugf("<%s> syncing to buffer", ch.Name)
	database.SyncTreeToBuffer(ch.NodeTree, ch.BufferDB)
	log.Debugf("<%s> tree synced to buffer", ch.Name)

	//ch.BufferDB.Print()

	// database.Cache represents bookmarks across all browsers
	// From browsers it should support: add/update
	// Delete method should only be possible through admin interface
	// We could have an @ignore command to ignore a bookmark

	// URLIndex is a hashmap index of all URLS representing current state
	// of the browser

	// nodeTree is current state of the browser as tree

	// Buffer is the current state of the browser represetned by
	// URLIndex and nodeTree

	// If the cache is empty just copy buffer to cache
	// until local db is already populated and preloaded
	//debugPrint("%d", BufferDB.Count())
	if empty, err := database.Cache.DB.IsEmpty(); empty {
		if err != nil {
			log.Error(err)
		}
		log.Info("cache empty: loading buffer to CacheDB")

		ch.BufferDB.CopyTo(database.Cache.DB)

		log.Debugf("syncing <%s> to disk", database.Cache.DB.Name)
	} else {
		ch.BufferDB.SyncTo(database.Cache.DB)
	}

	go database.Cache.DB.SyncToDisk(database.GetDBFullPath())
	ch.LastWatchRunTime = time.Since(startRun)

}

func NewChrome() *Chrome {
	return &Chrome{
		ChromeConfig: ChromeCfg,
	}
}

func init() {
	modules.RegisterBrowser(Chrome{ChromeConfig: ChromeCfg})
}
