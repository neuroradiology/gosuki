package logging

import (
	"os"

	"github.com/gin-gonic/gin"
	glogging "github.com/op/go-logging"
)

var log = glogging.MustGetLogger("MODE")

const ENV_GOMARK_MODE = "GOMARK_MODE"

const (
	DebugMode   string = "debug"
	ReleaseMode string = "release"
	TestMode    string = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

var gomarkMode = debugCode
var modeName = DebugMode

func SetMode(value string) {
	switch value {
	case DebugMode:
		gomarkMode = debugCode
	case ReleaseMode:
		gomarkMode = releaseCode
	case TestMode:
		gomarkMode = testCode
	default:
		log.Panic("go-bookmark mode unknown: " + value)
	}
	modeName = value
}

func RunMode() string {
	return modeName
}

func IsDebugging() bool {
	return gomarkMode == debugCode
}

func init() {
	mode := os.Getenv(ENV_GOMARK_MODE)
	if mode == "" {
		SetMode(DebugMode)
	} else {
		SetMode(mode)
		gin.SetMode(mode)
	}
}
