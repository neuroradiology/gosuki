package main

import (
	"gomark/logging"
)

var (
	// global logger
	log   = logging.GetLogger("")
	fflog = logging.GetLogger("FF")
)
