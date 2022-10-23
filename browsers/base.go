// ### Browsers API
//
// All constants and common API for browsers should be implemented here.
//
// For *browser specific* implementation create a new file for that browser.
//
// You must then implement a `func New[BrowserType]() IBrowser` function and
// implement parsing.
package browsers

import (
	"fmt"
	"io"
	"path"
	"path/filepath"
	"reflect"

	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/index"
	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/watch"

	"github.com/fsnotify/fsnotify"
	"github.com/sp4ke/hashmap"
)

var log = logging.GetLogger("BASE")

type BrowserType uint8

// Browser types
const (
	// Chromium based browsers (chrome, brave ... )
	TChrome BrowserType = iota

	// Firefox based browsers ie. they relay on places.sqlite
	TFirefox

	// Other
	TCustom
)

// Channel parameters
const eventsChanLen = 1000

// Used to store bookmark paths and other
// data related to a particular browser kind
//_TODO: replace in chrome with ProfileManager and remove this ref
type BrowserPaths struct {
	BookmarkFile string
	BookmarkDir  string
}

//FIX: in chome use Browser interface instead
type IBrowser interface {
	watch.Watchable

	Init() error // browser initializiation goes here
	RegisterHooks(...parsing.Hook)
	Load() error // Loads bookmarks to db without watching
	Shutdown()   // Graceful shutdown, it should call the BaseBrowser.Close()
}

// Base browser class serves as reference for implmented browser types
// Browser should contain enough data internally to not rely on any global
// variable or constant if possible.
// To create new browsers, you must implement a New<BrowserType>() IBrowser function
//
// `URLIndex` (HashMap RBTree):
// Used as fast query db representing the last known browser bookmarks.
//
// `nodeTree` (Tree DAG):
// Used in each job to represent bookmarks in a tree
//
// `BufferDB`: sqlite buffer used across jobs
type BaseBrowser struct {
	watcher    *watch.Watcher
	eventsChan chan fsnotify.Event

	// Dir where bookmarks file is stored
	BaseDir string
	// Name of bookmarks file
	BkFile string

	WatchedPaths []string

	// In memory sqlite db (named `memcache`).
	// Used to keep a browser's state of bookmarks across jobs.
	BufferDB *database.DB

	// Fast query db using an RB-Tree hashmap.
	// It represents a URL index of the last running job
	URLIndex *hashmap.RBTree

	// Pointer to the root of the node tree
	// The node tree is built again for every Run job on a browser
	NodeTree *tree.Node
	// Various parsing and timing stats
	Stats          *parsing.Stats
	Type           BrowserType
	Name           string
	IsWatching     bool
	UseFileWatcher bool
	parseHooks     []parsing.Hook

	io.Closer // Close database connections

	baseInit   bool
	bufferInit bool
}

func (bw *BaseBrowser) GetWatcher() *watch.Watcher {
	watcherType := reflect.TypeOf((*watch.Watcher)(nil)).Elem()
	// In case we use other types of watchers/events
	if reflect.TypeOf(bw.watcher) == reflect.PtrTo(watcherType) {
		return bw.watcher
	}
	return nil
}

func (bw *BaseBrowser) GetBookmarksPath() string {
	path, err := filepath.EvalSymlinks(path.Join(bw.BaseDir, bw.BkFile))
	if err != nil {
		log.Error(err)
	}
	return path
}

func (bw *BaseBrowser) InitEventsChan() {
	bw.eventsChan = make(chan fsnotify.Event, eventsChanLen)
}

func (bw *BaseBrowser) EventsChan() chan fsnotify.Event {
	return bw.eventsChan
}

func (bw *BaseBrowser) GetWatchedDir() string {
	return bw.BaseDir
}

// Setup file watcher using the provided []Watch elements
func (bw *BaseBrowser) SetupFileWatcher(watches ...*watch.Watch) {
	var err error

	if !bw.UseFileWatcher {
		return
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Critical(err)
	}

	watchedMap := make(map[string]*watch.Watch)
	for _, v := range watches {
		watchedMap[v.Path] = v
	}

	bw.watcher = &watch.Watcher{
		W:       fswatcher,
		Watched: watchedMap,
		Watches: watches,
	}

	// Add all watched paths
	for _, v := range watches {

		if err != nil {
			log.Critical(err)
		}
		err = bw.watcher.W.Add(v.Path)
		if err != nil {
			log.Critical(err)
		}
	}

}

func (bw *BaseBrowser) ResetWatcher() {
	err := bw.watcher.W.Close()
	if err != nil {
		log.Critical(err)
	}
	bw.SetupFileWatcher(bw.watcher.Watches...)
}

func (bw *BaseBrowser) Close() error {
	if bw.watcher != nil {
		err := bw.watcher.W.Close()
		if err != nil {
			return err
		}
	}

	if bw.bufferInit {
		err := bw.BufferDB.Close()
		if err != nil {
			return err
		}

	}

	return nil
}

func (b *BaseBrowser) Shutdown() {
	err := b.Close()
	if err != nil {
		log.Critical(err)
	}

	log.Debugf("<%s> shutdown complete ", b.Name)
}

// setup checks if Browser implements Initializer and Loader interfaces and
// call the appropriate methods
//_TODO: use BaseBrowser implementation for Init and Load !! 
//_TODO: so that implementers don't need to manually call the base Init / Load
func Setup(b Browser) error {
	log.Debugf("setting up <%s>", b)

	// Initialization logic
	if initializer, ok := b.(Initializer); ok {
		err := initializer.Init()
		if err != nil {
			return err
		}
	}

	// Loading logic
	if loader, ok := b.(Loader); ok {
		err := loader.Load()
		if err != nil {
			return err
		}

	}

	// var loader, ok := b.(Loader)

	return nil
}

//TODO: use the new interfaces Loader / Initializer 
func (b *BaseBrowser) Init() error {

	// Init browser buffer
	err := b.initBuffer()
	if err != nil {
		return err
	}
	b.bufferInit = true

	// Creates in memory Index (RB-Tree)
	b.URLIndex = index.NewIndex()

	b.baseInit = true

	return nil
}

//TODO: use the new interfaces Loader / Initializer 
func (bw *BaseBrowser) Load() error {

	if !bw.baseInit {
		return fmt.Errorf("base init on <%s> missing, call Init() on BaseBrowser", bw.Name)
	}

	// Check if cache is initialized
	if !database.Cache.IsInitialized() {
		return fmt.Errorf("<%s> Loading bookmarks while cache not yet initialized", bw.Name)
	}

	// In case we use other types of watchers/events
	if bw.UseFileWatcher && bw.watcher == nil {
		return fmt.Errorf("<%s> watcher not initialized, use SetupFileWatcher() when creating the browser", bw.Name)
	}

	log.Debugf("<%s> preloading bookmarks", bw.Name)

	return nil
}

func (b *BaseBrowser) RebuildIndex() {
	log.Debugf("<%s> rebuilding index based on current nodeTree", b.Name)
	b.URLIndex = index.NewIndex()
	tree.WalkBuildIndex(b.NodeTree, b.URLIndex)
}

func (b *BaseBrowser) RebuildNodeTree() {
	b.NodeTree = &tree.Node{
		Name:   "root",
		Parent: nil,
		Type:   "root",
	}
}

// init buffer db, TODO: defer closings and shutdown
func (b *BaseBrowser) initBuffer() error {
	var err error

	bufferName := fmt.Sprintf("buffer_%s", b.Name)
	b.BufferDB, err = database.New(bufferName, "", database.DBTypeInMemoryDSN).Init()
	if err != nil {
		return err
	}

	err = b.BufferDB.InitSchema()
	if err != nil {
		return err
	}

	return nil
}

func (b *BaseBrowser) RegisterHooks(hooks ...parsing.Hook) {
	log.Debugf("<%s> registering hooks", b.Name)
	for _, hook := range hooks {
		b.parseHooks = append(b.parseHooks, hook)
	}
}

func (b *BaseBrowser) ResetStats() {
	b.Stats.LastURLCount = b.Stats.CurrentUrlCount
	b.Stats.LastNodeCount = b.Stats.CurrentNodeCount
	b.Stats.CurrentNodeCount = 0
	b.Stats.CurrentUrlCount = 0
}

func (b *BaseBrowser) HasReducer() bool {
	return b.eventsChan != nil
}

func (b *BaseBrowser) String() string {
	return b.Name
}

// Runs browsed defined hooks on bookmark
func (b *BaseBrowser) RunParseHooks(node *tree.Node) {
	for _, hook := range b.parseHooks {
		hook(node)
	}
}
