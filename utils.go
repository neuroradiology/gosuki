package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// TODO
// Run debounce in it's own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func debounce(interval time.Duration, input chan fsnotify.Event, w IWatchable) {
	var item fsnotify.Event

	for {
		select {
		case item = <-input:
			log.Debugf("received an event %v on the spammy events channel", item.Op)
		case <-time.After(interval):
			log.Debug("Runngin parse method")
			w.Run()
		}
	}
}
