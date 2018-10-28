package main

import (
	"github.com/fsnotify/fsnotify"
)

// Used as input to WatcherThread
// It does not have to be a browser as long is the interface is implemented
type IWatchable interface {
	SetupWatcher()                 // Starts watching bookmarks and runs Load on change
	Watch() bool                   // starts watching linked watcher
	Run()                          // Callback fired on event
	GetWatcher() *fsnotify.Watcher // returns linked watcher
	ResetWatcher()                 // resets a new watcher
	GetPath() string               // returns watched path
	GetDir() string                // returns watched dir
	EventsChan() chan fsnotify.Event
}

// Main thread for watching file changes
func WatcherThread(w IWatchable) {

	bookmarkPath := w.GetPath()
	log.Infof("watching %s", bookmarkPath)

	for {
		// Keep watcher here as it is reset from within
		// the select block
		watcher := w.GetWatcher()

		select {
		case event := <-watcher.Events:

			// On Chrome like browsers the bookmarks file is created
			// at every change.

			/*
			 * When a file inside a watched directory is renamed/created,
			 * fsnotify does not seem to resume watching the newly created file, we
			 * need to destroy and create a new watcher. The ResetWatcher() and
			 * `break` statement ensure we get out of the `select` block and catch
			 * the newly created watcher to catch events even after rename/create
			 */

			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == bookmarkPath {

				w.Run()
				log.Debugf("event: %v | eventName: %v", event.Op, event.Name)

				log.Debugf("resetting watchers")
				w.ResetWatcher()

				break
			}

			// Firefox keeps the file open and makes changes on it
			// It needs a debouncer
			if event.Name == bookmarkPath {
				log.Debugf("event: %v | eventName: %v", event.Op, event.Name)
				//go debounce(1000*time.Millisecond, spammyEventsChannel, w)
				ch := w.EventsChan()
				ch <- event
				//w.Run()
			}
		case err := <-watcher.Errors:
			log.Error(err)
		}
	}
}
