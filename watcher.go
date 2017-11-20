package main

import (
	"github.com/fsnotify/fsnotify"
)

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
