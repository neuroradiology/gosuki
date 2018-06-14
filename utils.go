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
	var event fsnotify.Event
	var isResting bool
	timer := time.NewTimer(interval)

	for {
		select {
		case event = <-input:
			log.Debugf("received an event %v on the events channel", event.Op)

			if !isResting {
				// Run the job
				//log.Debug("Not resting, running job")
				w.Run()
				//log.Debug("Restting timer")
				timer.Reset(interval)
				//log.Debug("Is resting now")
				isResting = true
			}
			//else {
			//log.Debug("Resting, will not run job")
			//}

		case <-timer.C:
			//log.Debugf("timer done, not resting")
			isResting = false
		}
	}
}
