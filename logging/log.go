package logging

import (
	"os"

	glogging "github.com/op/go-logging"
)

type Logger = glogging.Logger

const (
	debugFmt   = `%{color} %{time:15:04:05.000} %{level:.4s} [%{module:.4s}] %{shortfunc:.10s}: %{color:reset} %{message}`
	releaseFmt = `[%{level}] - %{message}`
)

var (
	stdoutBackend         = glogging.NewLogBackend(os.Stderr, "", 0)
	debugFormat           = glogging.MustStringFormatter(debugFmt)
	releaseFormat         = glogging.MustStringFormatter(releaseFmt)
	debugBackendFormatter = glogging.NewBackendFormatter(stdoutBackend, debugFormat)
	backendFormatter      = glogging.NewBackendFormatter(stdoutBackend, releaseFormat)

	// Default logger
	log = glogging.MustGetLogger("")
)

func GetLogger(module string) *glogging.Logger {
	return glogging.MustGetLogger(module)
}

func InitLogDebug() {
	glogging.SetBackend(debugBackendFormatter)
}

func InitLog() {
	leveledBackend := glogging.AddModuleLevel(backendFormatter)
	leveledBackend.SetLevel(glogging.WARNING, "")
	glogging.SetBackend(leveledBackend)
}
