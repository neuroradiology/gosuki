package watch

import "time"

// Run reducer in its own thread when the watcher is started
// It receives a struct{event, func} and runs the func only once in the interval
func ReduceEvents(interval time.Duration,
				w WatchRunner) {
	watch := w.Watch()
	log.Debugf("starting reducer service for %s", watch.ID)

	eventsIn := w.Watch().eventsChan
	timer := time.NewTimer(interval)
	beat := time.NewTicker(1 * time.Second).C
	var events []bool

	for {
		select {
		case <-eventsIn:
			// log.Debug("[reducuer] received event, resetting watch interval !")
			timer.Reset(interval)
			events = append(events, true)

		case <-beat:
			// log.Debugf("reducer beat %s", watch.ID)

		case <-timer.C:
			if len(events) > 0 {
				log.Debug("<reduce>: calling Run()")
				w.Run()

				// Empty events queue
				events = make([]bool, 0)
			}
		case _, ok := <-eventsIn:
			if !ok {
				log.Warningf("Events channel closed for %s", watch.ID)
				return // Exit the function or handle the closed channel case
			}
		}
	}
}
