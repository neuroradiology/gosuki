package main

import (
	"os"

	logging "github.com/op/go-logging"
)

var (
	logBackend  *logging.LogBackend
	log         *logging.Logger
	debugFormat = logging.MustStringFormatter(
		`%{color}%{level:.4s} %{time:15:04:05.000} %{shortfunc:.10s}: %{color:reset} %{message}`,
	)
	releaseFormat = logging.MustStringFormatter(
		`[%{level}] - %{message}`,
	)
)

func initLogging() {

	log = logging.MustGetLogger("gomark")
	logBackend = logging.NewLogBackend(os.Stderr, "", 0)

	if IsDebugging() {
		debugBackendFormatter := logging.NewBackendFormatter(logBackend, debugFormat)
		logging.SetBackend(debugBackendFormatter)
	} else {
		backendFormatter := logging.NewBackendFormatter(logBackend, releaseFormat)
		leveledBackend := logging.AddModuleLevel(backendFormatter)
		leveledBackend.SetLevel(logging.WARNING, "")
		logging.SetBackend(leveledBackend)
	}
}

func IsDebugging() bool {
	return gomarkMode == debugCode
}

func debugPrint(format string, values ...interface{}) {
	log.Debugf("[debug] "+format, values...)
}
