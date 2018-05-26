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

// Chrome details
var Chrome = struct {
	BookmarkFile string
	BookmarkDir  string
}{
	"Bookmarks",
	"/home/spike/.config/google-chrome-unstable/Default/",
}

type IBrowser interface {
	IWatchable
	InitBuffer() // init buffer db, should be defered to close after call
	InitIndex()  // Creates in memory Index (RB-Tree)
	RegisterHooks(...ParseHook)
	Load() // Loads bookmarks to db without watching
	//Parse(...ParseHook) // Main parsing method with different parsing hooks
	Close() // Gracefully finish work and stop watching
}

// Base browser class serves as reference for implmented browser types
// Browser should contain enough data internally to not rely on any global
// variable or constant if possible.
// To create new browsers, you must implement a New<BrowserType>() IBrowser function
// URLIndex (HashMap RBTree):
// Used as fast query db representing the last known browser bookmarks.
// nodeTree (Tree DAG):
// Used as buffer. Represents current Browser bookmarks
type BaseBrowser struct {
	watcher    *fsnotify.Watcher
	baseDir    string
	bkFile     string
	bufferDB   *DB
	URLIndex   *hashmap.RBTree // Fast query of last browser state
	nodeTree   *Node           // pointer to root of the node tree
	stats      *ParserStats
	bType      BrowserType
	name       string
	isWatching bool
	parseHooks []ParseHook
}

func (bw *BaseBrowser) Watcher() *fsnotify.Watcher {
	return bw.watcher
}

func (bw *BaseBrowser) Load() {
	log.Debug("BaseBrowser Load()")
	bw.InitIndex()

	// Check if cache is initialized
	if cacheDB == nil || cacheDB.handle == nil {
		log.Critical("cache is not yet initialized !")
		panic("cache is not yet initialized !")
	}

	if bw.watcher == nil {
		log.Fatal("watcher not initialized, use SetupWatcher() when creating the browser !")
	}

	log.Debug("preloading bookmarks")
}

func (bw *BaseBrowser) GetPath() string {
	return path.Join(bw.baseDir, bw.bkFile)
}

func (bw *BaseBrowser) GetDir() string {
	return bw.baseDir
}

func (bw *BaseBrowser) SetupWatcher() {
	var err error
	bw.watcher, err = fsnotify.NewWatcher()
	logPanic(err)
	err = bw.watcher.Add(bw.baseDir)
	logPanic(err)
}

func (bw *BaseBrowser) Close() {
	err := bw.watcher.Close()
	bw.bufferDB.Close()
	logPanic(err)
}

func (b *BaseBrowser) InitIndex() {
	b.URLIndex = NewIndex()
}

func (b *BaseBrowser) RebuildIndex() {
	log.Debugf("Rebuilding index based on current nodeTree")
	b.URLIndex = NewIndex()
	WalkBuildIndex(b.nodeTree, b)
}

func (b *BaseBrowser) InitBuffer() {

	bufferName := fmt.Sprintf("buffer_%s", b.name)
	bufferPath := fmt.Sprintf(DBBufferFmt, bufferName)

	b.bufferDB = DB{}.New(bufferName, bufferPath)
	b.bufferDB.Init()

	b.bufferDB.Attach(cacheDB)
}

func (b *BaseBrowser) RegisterHooks(hooks ...ParseHook) {
	log.Debug("Registering hooks")
	for _, hook := range hooks {
		b.parseHooks = append(b.parseHooks, hook)
	}
}

// Runs browsed defined hooks on bookmark
func (b *BaseBrowser) RunParseHooks(node *Node) {
	for _, hook := range b.parseHooks {
		hook(node)
	}
}
