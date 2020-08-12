package main

import (
	"git.sp4ke.com/sp4ke/gomark/logging"
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
