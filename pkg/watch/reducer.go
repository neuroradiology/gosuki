package watch

import "time"

// Run reducer in its own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func ReduceEvents(interval time.Duration,
	w WatchRunner) {
	log.Debug("starting reducer service ...")

	eventsIn := w.Watch().eventsChan
	timer := time.NewTimer(interval)
	var events []bool

	for {
		select {
		case <-eventsIn:
			// log.Debug("[reducuer] received event, resetting watch interval !")
			timer.Reset(interval)
			events = append(events, true)

		case <-timer.C:
			if len(events) > 0 {
				log.Debug("<reduce>: running run event")
				w.Run()

				// Empty events queue
				events = make([]bool, 0)
			}
		}
	}
}
