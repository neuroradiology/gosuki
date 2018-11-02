package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// TODO
// Run reducer in it's own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func reducer(interval time.Duration, input chan fsnotify.Event, w IWatchable) {
	var waiting bool

	log.Debug("Running reducer")

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-input:
			log.Debugf("received event, len(chan):  %d ", len(input))

			if !waiting {
				waiting = true
				// Run the job
				log.Debug("Not resting")
				w.Run()

				//ticker = time.NewTicker(interval)
			} else { // Ignore this event
				log.Debug("resting")
				break
			}

		case <-ticker.C:
			//log.Debug("tick")
			ticker = time.NewTicker(interval)
			if waiting {
				waiting = false
			}
		}
	}
}
