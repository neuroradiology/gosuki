package logging

import (
	"fmt"
	"os"

	glogging "github.com/op/go-logging"
)

type Logger = glogging.Logger

const (
	debugDefaultFmt = `%{color} %{time:15:04:05.000} %{level:.4s} %{shortfunc:.10s}: %{color:reset} %{message}`
	debugFmt        = `%{color} %{time:15:04:05.000} %{level:.4s} [%{module:.4s}] %{shortfile}:%{shortfunc:.10s}: %{color:reset} %{message}`
	releaseFmt      = `[%{level}] - %{message}`
)

var (
	stdoutBackend         = glogging.NewLogBackend(os.Stderr, "", 0)
	debugFormatter        = glogging.MustStringFormatter(debugFmt)
	debugDefaultFormatter = glogging.MustStringFormatter(debugDefaultFmt)
	releaseFormatter      = glogging.MustStringFormatter(releaseFmt)

	debugBackend        = glogging.NewBackendFormatter(stdoutBackend, debugFormatter)
	debugDefaultBackend = glogging.NewBackendFormatter(stdoutBackend, debugDefaultFormatter)
	releaseBackend      = glogging.NewBackendFormatter(stdoutBackend, releaseFormatter)

	debugMode   bool
	loggers     map[string]*glogging.Logger
	usedLoggers map[string]bool

	// Default debug leveledBacked
	leveledDefaultDebug = glogging.AddModuleLevel(debugDefaultBackend)
	leveledDebug        = glogging.AddModuleLevel(debugBackend)
	leveledRelease      = glogging.AddModuleLevel(releaseBackend)
)

// Register which loggers to use
func UseLogger(module string) {
	usedLoggers[module] = true
}

func GetLogger(module string) *glogging.Logger {
	logger := glogging.MustGetLogger(module)
	if len(module) > 0 {
		loggers[module] = logger
	} else {
		loggers["default"] = logger
	}

	if debugMode {
		if len(module) > 0 {
			logger.SetBackend(leveledDebug)
		} else {
			logger.SetBackend(leveledDefaultDebug)
		}
	} else {
		logger.SetBackend(leveledRelease)
	}

	return logger
}

func SetDebug(d bool) {
	debugMode = d

	if !debugMode {
		for m, log := range loggers {
			fmt.Println(m)
			log.SetBackend(leveledRelease)
		}
	}
}

func init() {
	debugMode = IsDebugging()
	// init global vars
	loggers = make(map[string]*glogging.Logger)
	// Sets the default backend for all new loggers
	glogging.SetBackend(debugDefaultBackend)

	// Release level
	leveledRelease.SetLevel(glogging.WARNING, "")
}
