package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// TODO
// Run debounce in it's own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func debouncer(interval time.Duration, input chan fsnotify.Event, w IWatchable) {
	log.Debug("Running debouncer")
	//var event fsnotify.Event

	ticker := time.NewTicker(interval)

	for {
		select {
		//case event = <-input:
		//log.Debugf("received an event %v on the spammy events channel", event.Op)

		//// Run the job
		////w.Run()

		case <-ticker.C:
			log.Debugf("debouncer ticker ... events len: %v", len(input))
			log.Debug("implement a queue ! Should not use channels as queues")
		}
	}
}
