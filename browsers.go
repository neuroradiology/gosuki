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
	watcher    *fsnotify.Watcher
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
	Stats      *ParserStats
	bType      BrowserType
	name       string
	isWatching bool
	parseHooks []ParseHook
}

func (bw *BaseBrowser) GetWatcher() *fsnotify.Watcher {
	return bw.watcher
}

func (bw *BaseBrowser) Load() {
	log.Debug("BaseBrowser Load()")
	bw.InitIndex()

	// Check if cache is initialized
	if CacheDB == nil || CacheDB.Handle == nil {
		log.Criticalf("<%s> Loading bookmarks while cache not yet initialized !", bw.name)
	}

	if bw.watcher == nil {
		log.Fatal("watcher not initialized, use SetupWatcher() when creating the browser !")
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

func (bw *BaseBrowser) SetupWatcher() {
	var err error
	bw.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Critical(err)
	}
	err = bw.watcher.Add(bw.GetDir())
	if err != nil {
		log.Critical(err)
	}

}

func (bw *BaseBrowser) ResetWatcher() {
	err := bw.watcher.Close()
	if err != nil {
		log.Critical(err)
	}
	bw.SetupWatcher()
}

func (bw *BaseBrowser) Close() error {
	err := bw.watcher.Close()
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

// Runs browsed defined hooks on bookmark
func (b *BaseBrowser) RunParseHooks(node *Node) {
	for _, hook := range b.parseHooks {
		hook(node)
	}
}
