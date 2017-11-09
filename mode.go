package main

import "os"

const ENV_GO_BOOKMARK_MODE = "GO_BOOKMARK_MODE"

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

var goBookmarkMode = debugCode
var modeName = DebugMode

func init() {
	mode := os.Getenv(ENV_GO_BOOKMARK_MODE)
	if mode == "" {
		SetMode(DebugMode)
	} else {
		SetMode(mode)
	}
}

func SetMode(value string) {
	switch value {
	case DebugMode:
		goBookmarkMode = debugCode
	case ReleaseMode:
		goBookmarkMode = releaseCode
	case TestMode:
		goBookmarkMode = testCode
	default:
		panic("go-bookmark mode unknown: " + value)
	}
	modeName = value
}

func Mode() string {
	return modeName
}
