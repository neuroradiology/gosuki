package main

import (
	"database/sql"
	"log"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
)

type bMarkTypes int

const (
	Chrome bMarkTypes = iota
	Firefox
)

type bookmarkWatcher struct {
	watcher   *fsnotify.Watcher
	baseDir   string
	bkFile    string
	parseFunc func(*bookmarkWatcher)
	bufferDB  *sql.DB
}

func (bw *bookmarkWatcher) Close() error {
	if err := bw.watcher.Close(); err != nil {
		return err
	}

	return nil
}

func (bw *bookmarkWatcher) Init(basedir string, bkfile string, bkType bMarkTypes) *bookmarkWatcher {
	var err error

	bw.baseDir = basedir
	bw.bkFile = bkfile

	bw.watcher, err = fsnotify.NewWatcher()
	logPanic(err)

	switch bkType {
	case Chrome:
		bw.parseFunc = googleParseBookmarks
	}

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
	debugPrint("watching %s", bookmarkPath)

	for {
		select {
		case event := <-bw.watcher.Events:

			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == bookmarkPath {

				debugPrint("event: %v | eventName: %v", event.Op, event.Name)
				debugPrint("modified file: %s", event.Name)
				start := time.Now()
				parseFunc(bw)
				elapsed := time.Since(start)
				debugPrint("parsed in %s", elapsed)
				debugPrint("%v", _sql3conns)

			}
		case err := <-bw.watcher.Errors:
			log.Println("error:", err)
		}
	}

	debugPrint("Exiting watch thread")
}
