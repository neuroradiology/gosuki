package main

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

// TODO: Use `interval` countdown since last event, if count down reaches 0
// execute one event from channel and discard all the rest

// Run reducer in it's own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func reducer(interval time.Duration, input chan fsnotify.Event, w IWatchable) {

	log.Debug("Running reducer")

	timer := time.NewTimer(interval)
	//lastRun := time.Now()
	var events []bool

	for {
		select {
		case <-input:
			timer.Reset(interval)
			events = append(events, true)
			//log.Debugf("received event, len(chan):  %d ", len(input))

			// Run is executed once every `interval`time
			//if time.Since(lastRun) > interval {
			//w.Run()
			//lastRun = time.Now()

			//} else { // dispose of this event
			////log.Debugf("discarding, chan len %d", len(input))
			//break
			//}
		case <-timer.C:
			if len(events) > 0 {
				w.Run()

				// Empty events queue
				events = make([]bool, 0)
			}
		}
	}
}
