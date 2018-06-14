package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

var Firefox = BrowserPaths{
	"places.sqlite",
	"/home/spike/.mozilla/firefox/p1rrgord.default/",
}

type FFBrowser struct {
	BaseBrowser //embedding
}

func NewFFBrowser() IBrowser {
	browser := &FFBrowser{}
	browser.name = "firefox"
	browser.bType = TFirefox
	browser.baseDir = Firefox.BookmarkDir
	browser.bkFile = Firefox.BookmarkFile
	browser.Stats = &ParserStats{}
	browser.NodeTree = &Node{Name: "root", Parent: nil}
	browser.eventsChan = make(chan fsnotify.Event, EventsChanLen)

	// Across jobs buffer
	browser.InitBuffer()

	browser.SetupWatcher()

	// Run debouncer to avoid duplicate running of jobs
	// when a batch of events is received
	go debouncer(3000*time.Millisecond, browser.eventsChan, browser)

	return browser
}

func (bw *FFBrowser) Watch() bool {

	if !bw.isWatching {
		go WatcherThread(bw)
		bw.isWatching = true
		return true
	}

	return false
}

func (bw *FFBrowser) Load() {
	bw.BaseBrowser.Load()
	bw.Run()
}

func (bw *FFBrowser) Run() {

	log.Infof("Parsing Firefox bookmarks\n")

}
