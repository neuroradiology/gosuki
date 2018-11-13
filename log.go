package main

import (
	"gomark/logging"
)

var (
	// global logger
	log   = logging.GetLogger("")
	fflog = logging.GetLogger("FF")
)

func init() {
	//logging.SetLogger("FF", logging.WARNING)
	//logging.UseLogger("STATS", nil)
}
