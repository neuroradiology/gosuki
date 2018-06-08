package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// Used as input to WatcherThread
// It does not have to be a browser as long is the interface is implemented
type IWatchable interface {
	SetupWatcher()              // Starts watching bookmarks and runs Load on change
	Watch() bool                // starts watching linked watcher
	Run()                       // Callback fired on event
	Watcher() *fsnotify.Watcher // returns linked watcher
	GetPath() string            // returns watched path
	GetDir() string             // returns watched dir
}

// Main thread for watching file changes
func WatcherThread(w IWatchable) {

	spammyEventsChannel := make(chan fsnotify.Event)

	bookmarkPath := w.GetPath()
	log.Infof("watching %s", bookmarkPath)

	watcher := w.Watcher()

	for {
		select {
		case event := <-watcher.Events:

			// On Chrome like browsers the bookmarks file is created
			// at every change.
			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == bookmarkPath {

				debugPrint("event: %v | eventName: %v", event.Op, event.Name)
				//debugPrint("modified file: %s", event.Name)
				//start := time.Now()
				//parseFunc(bw)
				w.Run()
				//elapsed := time.Since(start)
				//debugPrint("parsed in %s", elapsed)
			}

			// Firefox keeps the file open and makes changes on it
			if event.Op&fsnotify.Write == fsnotify.Write &&
				event.Name == bookmarkPath {
				debugPrint("event: %v | eventName: %v", event.Op, event.Name)
				go debounce(1000*time.Millisecond, spammyEventsChannel, w)
				spammyEventsChannel <- event
				//w.Run()
			}
		case err := <-watcher.Errors:
			log.Errorf("error: %s", err)
		}
	}
}
