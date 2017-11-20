package main

import (
	"github.com/fsnotify/fsnotify"
)

type IWatchable interface {
	Watch() bool
	Watcher() *fsnotify.Watcher // returns linked watcher
	Parse()                     // Main parsing method
	GetPath() string            // returns watched path
	GetDir() string             // returns watched dir
}

func WatcherThread(w IWatchable) {

	bookmarkPath := w.GetPath()
	log.Infof("watching %s", bookmarkPath)

	watcher := w.Watcher()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == bookmarkPath {

				debugPrint("event: %v | eventName: %v", event.Op, event.Name)
				//debugPrint("modified file: %s", event.Name)
				//start := time.Now()
				//parseFunc(bw)
				w.Parse()
				//elapsed := time.Since(start)
				//debugPrint("parsed in %s", elapsed)
			}
		case err := <-watcher.Errors:
			log.Errorf("error: %s", err)
		}
	}
}
