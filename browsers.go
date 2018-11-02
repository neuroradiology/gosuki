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
	"path"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/sp4ke/hashmap"
)

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
	InitBuffer() // init buffer db, TODO: defer closings and shutdown
	InitIndex()  // Creates in memory Index (RB-Tree)
	RegisterHooks(...ParseHook)
	Load() // Loads bookmarks to db without watching
	//Parse(...ParseHook) // Main parsing method with different parsing hooks
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
	NodeTree *Node

	// Various parsing and timing stats
	Stats          *ParserStats
	bType          BrowserType
	name           string
	isWatching     bool
	useFileWatcher bool
	parseHooks     []ParseHook
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
	if CacheDB == nil || CacheDB.handle == nil {
		log.Criticalf("<%s> Loading bookmarks while cache not yet initialized !", bw.name)
	}

	// In case we use other types of watchers/events
	if bw.useFileWatcher && bw.watcher == nil {
		log.Warningf("<%s> watcher not initialized, use SetupFileWatcher() when creating the browser !", bw.name)
	}

	log.Debugf("<%s> preloading bookmarks", bw.name)
}

func (bw *BaseBrowser) GetPath() string {
	return path.Join(bw.baseDir, bw.bkFile)
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
		watchedMap[v.path] = v
	}

	bw.watcher = &Watcher{
		w:       fswatcher,
		watched: watchedMap,
		watches: watches,
	}

	// Add all watched paths
	for _, v := range watches {

		err = bw.watcher.w.Add(v.path)
		if err != nil {
			log.Critical(err)
		}
	}

}

func (bw *BaseBrowser) ResetWatcher() {
	err := bw.watcher.w.Close()
	if err != nil {
		log.Critical(err)
	}
	bw.SetupFileWatcher(bw.watcher.watches...)
}

func (bw *BaseBrowser) Close() error {
	err := bw.watcher.w.Close()
	if err != nil {
		return err
	}

	err = bw.BufferDB.Close()
	if err != nil {
		return err
	}

	return nil
}

func (b *BaseBrowser) InitIndex() {
	b.URLIndex = NewIndex()
}

func (b *BaseBrowser) RebuildIndex() {
	log.Debugf("Rebuilding index based on current nodeTree")
	b.URLIndex = NewIndex()
	WalkBuildIndex(b.NodeTree, b)
}

func (b *BaseBrowser) RebuildNodeTree() {
	b.NodeTree = &Node{Name: "root", Parent: nil, Type: "root"}
}

func (b *BaseBrowser) InitBuffer() {

	bufferName := fmt.Sprintf("buffer_%s", b.name)
	bufferPath := fmt.Sprintf(DBBufferFmt, bufferName)

	b.BufferDB = DB{}.New(bufferName, bufferPath)
	b.BufferDB.Init()

	b.BufferDB.Attach(CacheDB)
}

func (b *BaseBrowser) RegisterHooks(hooks ...ParseHook) {
	log.Debug("Registering hooks")
	for _, hook := range hooks {
		b.parseHooks = append(b.parseHooks, hook)
	}
}

func (b *BaseBrowser) ResetStats() {
	b.Stats.lastURLCount = b.Stats.currentUrlCount
	b.Stats.lastNodeCount = b.Stats.currentNodeCount
	b.Stats.currentNodeCount = 0
	b.Stats.currentUrlCount = 0
}

func (b *BaseBrowser) HasReducer() bool {
	return b.eventsChan != nil
}

func (b *BaseBrowser) Name() string {
	return b.name
}

// Runs browsed defined hooks on bookmark
func (b *BaseBrowser) RunParseHooks(node *Node) {
	for _, hook := range b.parseHooks {
		hook(node)
	}
}
