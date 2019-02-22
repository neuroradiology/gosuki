// ### Browsers API
//
// All constants and common API for browsers should be implemented here.
//
// For *browser specific* implementation create a new file for that browser.
//
// You must then implement a `func New[BrowserType]() IBrowser` function and
// implement parsing.
package main

import (
	"fmt"
	"gomark/database"
	"gomark/index"
	"gomark/parsing"
	"gomark/tree"
	"gomark/watch"
	"io"
	"path"
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/sp4ke/hashmap"
)

type IWatchable = watch.Watchable
type Watcher = watch.Watcher
type Watch = watch.Watch

type BrowserType uint8

// Browser types
const (
	TChrome BrowserType = iota
	TFirefox
)

// Channel parameters
const EventsChanLen = 1000

// Used to store bookmark paths and other
// data related to a particular browser kind
type BrowserPaths struct {
	BookmarkFile string
	BookmarkDir  string
}

type IBrowser interface {
	IWatchable
	InitBuffer() error // init buffer db, TODO: defer closings and shutdown
	InitIndex()        // Creates in memory Index (RB-Tree)
	RegisterHooks(...parsing.Hook)
	Load()     // Loads bookmarks to db without watching
	Shutdown() // Graceful shutdown, it should call the BaseBrowser.Close()
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
	watcher    *Watcher
	eventsChan chan fsnotify.Event
	baseDir    string
	bkFile     string

	// In memory sqlite db (named `memcache`).
	// Used to keep a browser's state of bookmarks across jobs.
	BufferDB *DB

	// Fast query db using an RB-Tree hashmap.
	// It represents a URL index of the last running job
	URLIndex *hashmap.RBTree

	// Pointer to the root of the node tree
	// The node tree is built again for every Run job on a browser
	NodeTree *tree.Node
	// Various parsing and timing stats
	Stats          *parsing.Stats
	bType          BrowserType
	name           string
	isWatching     bool
	useFileWatcher bool
	parseHooks     []parsing.Hook

	io.Closer // Close database connections
}

func (bw *BaseBrowser) GetWatcher() *Watcher {
	watcherType := reflect.TypeOf((*Watcher)(nil)).Elem()
	// In case we use other types of watchers/events
	if reflect.TypeOf(bw.watcher) == reflect.PtrTo(watcherType) {
		return bw.watcher
	}
	return nil
}
func (bw *BaseBrowser) Load() {
	log.Debug("BaseBrowser Load()")

	bw.InitIndex()

	// Check if cache is initialized
	if CacheDB == nil || CacheDB.Handle == nil {
		log.Criticalf("<%s> Loading bookmarks while cache not yet initialized !", bw.name)
	}

	// In case we use other types of watchers/events
	if bw.useFileWatcher && bw.watcher == nil {
		log.Warningf("<%s> watcher not initialized, use SetupFileWatcher() when creating the browser !", bw.name)
	}

	log.Debugf("<%s> preloading bookmarks", bw.name)
}

func (bw *BaseBrowser) GetBookmarksPath() string {
	path, err := filepath.EvalSymlinks(path.Join(bw.baseDir, bw.bkFile))
	if err != nil {
		log.Error(err)
	}
	return path
}

func (bw *BaseBrowser) EventsChan() chan fsnotify.Event {
	return bw.eventsChan
}

func (bw *BaseBrowser) GetDir() string {
	return bw.baseDir
}

// Setup file watcher using the provided []Watch elements
func (bw *BaseBrowser) SetupFileWatcher(watches ...*Watch) {
	var err error

	if !bw.useFileWatcher {
		return
	}

	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Critical(err)
	}

	watchedMap := make(map[string]*Watch)
	for _, v := range watches {
		watchedMap[v.Path] = v
	}

	bw.watcher = &Watcher{
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
	err := bw.watcher.W.Close()
	if err != nil {
		return err
	}

	err = bw.BufferDB.Close()
	if err != nil {
		return err
	}

	return nil
}

func (bw *BaseBrowser) Shutdown() {
	err := bw.Close()
	if err != nil {
		log.Critical(err)
	}
}

func (b *BaseBrowser) InitIndex() {
	b.URLIndex = index.NewIndex()
}

func (b *BaseBrowser) RebuildIndex() {
	log.Debugf("Rebuilding index based on current nodeTree")
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

func (b *BaseBrowser) InitBuffer() error {
	var err error

	bufferName := fmt.Sprintf("buffer_%s", b.name)
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
	log.Debug("Registering hooks")
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
	return b.name
}

// Runs browsed defined hooks on bookmark
func (b *BaseBrowser) RunParseHooks(node *tree.Node) {
	for _, hook := range b.parseHooks {
		hook(node)
	}
}
