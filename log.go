package main

import "gomark/logging"

var (
	// global logger
	log   = logging.GetLogger("")
	fflog = logging.GetLogger("FF")
)

func init() {
	if IsDebugging() {
		logging.InitLogDebug()
	} else {
		logging.InitLog()
	}
}
