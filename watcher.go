package main

import (
	"github.com/fsnotify/fsnotify"
)

// Used as input to WatcherThread
// It does not have to be a browser as long is the interface is implemented
type IWatchable interface {
	Name() string               // Name of the watchable
	HasReducer() bool           // Does the watchable has a reducer
	SetupFileWatcher(...*Watch) // Starts watching bookmarks and runs Load on change
	Watch() bool                // starts watching linked watcher
	Run()                       // Callback fired on event
	GetWatcher() *Watcher       // returns linked watcher
	ResetWatcher()              // resets a new watcher
	GetPath() string            // returns watched path
	GetDir() string             // returns watched dir
	EventsChan() chan fsnotify.Event
}

// Wrapper around fsnotify watcher
type Watcher struct {
	w       *fsnotify.Watcher // underlying fsnotify watcher
	watched map[string]*Watch // watched paths
	watches []*Watch          // helper var
}

// Details about the object being watched
type Watch struct {
	path       string        // Path to watch for events
	eventTypes []fsnotify.Op // events to watch for
	eventNames []string      // event names to watch for (file/dir names)
	resetWatch bool          // Reset the watcher when the event happens (useful for create events)
}

// Main thread for watching file changes
func WatcherThread(w IWatchable) {

	log.Infof("<%s> Started watcher", w.Name())
	for {
		// Keep watcher here as it is reset from within
		// the select block
		watcher := w.GetWatcher()
		resetWatch := false

		select {
		case event := <-watcher.w.Events:

			// Very verbose
			log.Debugf("event: %v | eventName: %v", event.Op, event.Name)

			// On Chrome like browsers the bookmarks file is created
			// at every change.

			/*
			 * When a file inside a watched directory is renamed/created,
			 * fsnotify does not seem to resume watching the newly created file, we
			 * need to destroy and create a new watcher. The ResetWatcher() and
			 * `break` statement ensure we get out of the `select` block and catch
			 * the newly created watcher to catch events even after rename/create
			 */

			for _, watched := range watcher.watches {
				for _, watchedEv := range watched.eventTypes {
					for _, watchedName := range watched.eventNames {

						if event.Op&watchedEv == watchedEv &&
							event.Name == watchedName {

							// For watchers who need a reducer
							// to avoid spammy events
							if w.HasReducer() {
								ch := w.EventsChan()
								ch <- event
							} else {
								//w.Run()
							}

							//log.Warning("event: %v | eventName: %v", event.Op, event.Name)

							if watched.resetWatch {
								log.Debugf("resetting watchers")
								w.ResetWatcher()
								resetWatch = true // needed to break out of big loop
							}

						}
					}
				}
			}

			if resetWatch {
				break
			}

			// Firefox keeps the file open and makes changes on it
			// It needs a debouncer
			//if event.Name == bookmarkPath {
			//log.Debugf("event: %v | eventName: %v", event.Op, event.Name)
			////go debounce(1000*time.Millisecond, spammyEventsChannel, w)
			//ch := w.EventsChan()
			//ch <- event
			////w.Run()
			//}
		case err := <-watcher.w.Errors:
			log.Error(err)
		}
	}
}
