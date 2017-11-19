package main

import (
	"database/sql"

	fsnotify "gopkg.in/fsnotify.v1"
)

type BrowserType uint8

const (
	TChromeBrowser BrowserType = iota
	FirefoxBrowser
)

type Browser interface {
	New(BrowserType) *Browser // Creates and initialize new browser
	Watch() *fsnotify.Watcher // Starts watching bookmarks and runs Load on change
	Load()                    // Loads bookmarks to db without watching
	Parse()                   // Main parsing method
	//Parse(...ParseHook) // Main parsing method with different parsing hooks
	Close() // Gracefully finish work and stop watching
}

// Base browser class serves as reference for implmented browser types
// Browser should contain enough data internally to not rely on any global
// variable or constant if possible.

type BaseBrowser struct {
	watcher   *fsnotify.Watcher
	baseDir   string
	bkFile    string
	parseFunc func(*Browser)
	bufferDB  *sql.DB
	stats     *parserStats
}

type ChromeBrowser struct {
	BaseBrowser //embedding
}
