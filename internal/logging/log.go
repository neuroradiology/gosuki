//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

package logging

import (
	"os"

	glogging "github.com/op/go-logging"
)

type Logger = glogging.Logger

const (
	debugDefaultFmt = `%{color} %{time:15:04:05.000} %{level:.4s} %{shortfunc:.10s}: %{color:reset} %{message}`
	debugFmt        = `%{color} %{time:15:04:05.000} %{level:.4s} [%{module:.4s}] %{shortfile}:%{shortfunc:.10s}: %{color:reset} %{message}`
	releaseFmt      = `%{color}[%{level:.4s}]%{color:reset} %{message}`
)

var (
	stdoutBackend = glogging.NewLogBackend(os.Stderr, "", 0)
	nullBackend   = glogging.NewLogBackend(new(NullWriter), "", 0)

	debugFormatter        = glogging.MustStringFormatter(debugFmt)
	debugDefaultFormatter = glogging.MustStringFormatter(debugDefaultFmt)
	releaseFormatter      = glogging.MustStringFormatter(releaseFmt)

	debugBackend        = glogging.NewBackendFormatter(stdoutBackend, debugFormatter)
	debugDefaultBackend = glogging.NewBackendFormatter(stdoutBackend, debugDefaultFormatter)
	releaseBackend      = glogging.NewBackendFormatter(stdoutBackend, releaseFormatter)
	silentBackend      = glogging.NewBackendFormatter(nullBackend, debugDefaultFormatter)

	loggers map[string]*glogging.Logger

	// Default debug leveledBacked
	leveledDefaultDebug = glogging.AddModuleLevel(debugDefaultBackend)
	leveledDebug        = glogging.AddModuleLevel(debugBackend)
	leveledRelease      = glogging.AddModuleLevel(releaseBackend)
    leveledSilent = glogging.AddModuleLevel(silentBackend)

	LoggingLevels = map[int]int{
		Release: int(glogging.WARNING),
		Info:    int(glogging.INFO),
		Debug:   int(glogging.DEBUG),
	}
)

type NullWriter struct{}

func (nw *NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func GetLogger(module string) *glogging.Logger {
	logger := glogging.MustGetLogger(module)
	if len(module) > 0 {
		loggers[module] = logger
	} else {
		loggers["default"] = logger
	}

	if LoggingMode >= Info {
		// fmt.Println("setting backend to >= info")
		if len(module) > 0 {
			logger.SetBackend(leveledDebug)
		} else {
			logger.SetBackend(leveledDefaultDebug)
		}
	} else {
		// fmt.Println("setting backend to release")
		logger.SetBackend(leveledRelease)
	}

	// setting log level for logger
	glogging.SetLevel(glogging.Level(LoggingLevels[LoggingMode]), module)

	// Register which loggers to use
	return logger
}

func setLogLevel(lvl int) {
	for k, logger := range loggers {
		// fmt.Printf("setting log level to:%v for %v\n ", LoggingLevels[lvl], k)
		glogging.SetLevel(glogging.Level(LoggingLevels[lvl]), k)

		if lvl >= Info {
			// fmt.Println("setting backend to debug for ", k)
			logger.SetBackend(leveledDebug)
		} else if lvl == -1 {
            logger.SetBackend(leveledSilent)
        } else {
			logger.SetBackend(leveledRelease)
			// fmt.Println("setting backend to release for ", k)
		}
	}
}

//FIX: Suppress output during testing

func init() {
	initRuntimeMode()

	// init global vars
	loggers = make(map[string]*glogging.Logger)

	// Sets the default backend for all new loggers
	//RELEASE: set to release when app released
	glogging.SetBackend(debugBackend)

	// Release level
	leveledRelease.SetLevel(glogging.WARNING, "")
}
