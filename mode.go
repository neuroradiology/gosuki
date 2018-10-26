package main

import (
	"os"
)

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

func initMode() {
	mode := os.Getenv(ENV_GOMARK_MODE)
	if mode == "" {
		SetMode(DebugMode)
	} else {
		SetMode(mode)
	}
}

func SetMode(value string) {
	switch value {
	case DebugMode:
		gomarkMode = debugCode
	case ReleaseMode:
		gomarkMode = releaseCode
	case TestMode:
		gomarkMode = testCode
	default:
		panic("go-bookmark mode unknown: " + value)
	}
	modeName = value
}

func RunMode() string {
	return modeName
}
