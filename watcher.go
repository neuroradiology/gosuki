package main

import (
	"database/sql"
	"path"

	"github.com/fsnotify/fsnotify"
)

type BrowserType int

const (
	Chrome BrowserType = iota
	Firefox
)

type bookmarkWatcher struct {
	watcher   *fsnotify.Watcher
	baseDir   string
	bkFile    string
	parseFunc func(*bookmarkWatcher)
	bufferDB  *sql.DB
	stats     *parserStat
}

func (bw *bookmarkWatcher) Close() error {
	if err := bw.watcher.Close(); err != nil {
		return err
	}

	return nil
}

func (bw *bookmarkWatcher) Init(basedir string, bkfile string, bkType BrowserType) *bookmarkWatcher {
	var err error

	bw.baseDir = basedir
	bw.bkFile = bkfile

	bw.stats = &parserStat{}

	bw.watcher, err = fsnotify.NewWatcher()
	logPanic(err)

	switch bkType {
	case Chrome:
		bw.parseFunc = googleParseBookmarks
	}

	return bw
}

func (bw *bookmarkWatcher) Preload() *bookmarkWatcher {

	// Check if cache is initialized
	if cacheDB == nil || cacheDB.handle == nil {
		log.Critical("cache is not yet initialized !")
		panic("cache is not yet initialized !")
	}

	if bw.watcher == nil {
		log.Fatal("please run bookmarkWatcher.Init() first !")
	}

	debugPrint("preloading bookmarks")
	bw.parseFunc(bw)

	return bw
}

func (bw *bookmarkWatcher) Start() error {

	if err := bw.watcher.Add(bw.baseDir); err != nil {
		return err
	}

	go bWatcherThread(bw, bw.parseFunc)

	return nil
}

func bWatcherThread(bw *bookmarkWatcher, parseFunc func(bw *bookmarkWatcher)) {

	bookmarkPath := path.Join(bw.baseDir, bw.bkFile)
	log.Infof("watching %s", bookmarkPath)

	for {
		select {
		case event := <-bw.watcher.Events:

			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == bookmarkPath {

				debugPrint("event: %v | eventName: %v", event.Op, event.Name)
				//debugPrint("modified file: %s", event.Name)
				//start := time.Now()
				parseFunc(bw)
				//elapsed := time.Since(start)
				//debugPrint("parsed in %s", elapsed)
			}
		case err := <-bw.watcher.Errors:
			log.Errorf("error: %s", err)
		}
	}

	debugPrint("Exiting watch thread")
}
