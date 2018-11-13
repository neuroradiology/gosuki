package tools

import (
	"gomark/logging"
	"gomark/watch"
	"time"

	"github.com/fsnotify/fsnotify"
)

var log = logging.GetLogger("WATCH")

// Run reducer in its own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func ReduceEvents(interval time.Duration, input chan fsnotify.Event, w watch.IWatchable) {
	log.Debug("Running reducer")

	timer := time.NewTimer(interval)
	var events []bool

	for {
		select {
		case <-input:
			timer.Reset(interval)
			events = append(events, true)

		case <-timer.C:
			if len(events) > 0 {
				w.Run()

				// Empty events queue
				events = make([]bool, 0)
			}
		}
	}
}
