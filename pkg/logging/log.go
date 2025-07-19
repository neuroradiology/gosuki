//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
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
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	log "github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

const EnvGosukiDebug = "GOSUKI_DEBUG"

const (
	Silent = -1 + iota
	Release
	Dev
)

const (
	debugDefaultFmt = `%{color} %{time:15:04:05.000} %{level:.4s} %{shortfunc:.10s}: %{color:reset} %{message}`
	debugFmt        = `%{color} %{time:15:04:05.000} %{level:.4s} [%{module:.4s}] %{shortfile}:%{shortfunc:.10s}: %{color:reset} %{message}`
	releaseFmt      = `%{color}[%{level:.4s}]%{color:reset} %{message}`
)

var (
	//RELEASE: Change to Release for release mode
	LoggingMode = Release
	TUIMode     bool
)

var (
	DefaultLogLevels = map[int]int{
		Release: 1,
		Dev:     3,
	}

	// Map cli log level to log.Level
	LogLvlMap = map[int]log.Level{
		-1: math.MaxInt32,
		0:  log.ErrorLevel,
		1:  log.WarnLevel,
		2:  log.InfoLevel,
		3:  log.DebugLevel,
	}

	loggers map[string]*log.Logger

	logTextStyle = lipgloss.NewStyle().Foreground(
		lipgloss.AdaptiveColor{Light: "245", Dark: "252"},
	)
	logTextFaintStyle = lipgloss.NewStyle().Foreground(
		lipgloss.AdaptiveColor{Light: "240", Dark: "246"},
	)
	logLevelStyles = map[log.Level]lipgloss.Style{
		log.DebugLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.DebugLevel.String())).
			MaxWidth(4).
			Foreground(lipgloss.Color("63")),
		log.InfoLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.InfoLevel.String())).
			// Bold(true).
			MaxWidth(4).
			Foreground(lipgloss.Color("36")),
		log.WarnLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.WarnLevel.String())).
			MaxWidth(4).
			Foreground(lipgloss.Color("178")),
		log.ErrorLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.ErrorLevel.String())).
			MaxWidth(4).
			Foreground(lipgloss.Color("204")),
		log.FatalLevel: lipgloss.NewStyle().
			SetString(strings.ToUpper(log.FatalLevel.String())).
			MaxWidth(4).
			Foreground(lipgloss.Color("134")),
	}
)

func GetLogger(module string) *log.Logger {
	if LoggingMode == Silent {
		return log.New(io.Discard)
	}

	lg := log.New(os.Stdout)

	if LoggingMode == Dev {
		lg.SetPrefix(fmt.Sprintf("[%.4s]", strings.ToUpper(module)))
		lg.SetTimeFormat(time.TimeOnly)
		lg.SetReportTimestamp(true)
		lg.SetCallerFormatter(func(file string, line int, _ string) string {
			return fmt.Sprintf("%s:%d", trimCallerPath(file, 1), line)
		})
		lg.SetReportCaller(true)
		lg.SetLevel(log.DebugLevel)

		//RELEASE:
	}

	loggers[module] = lg

	return lg
}

func newReleaseLogger(module string) *log.Logger {
	if len(module) > 0 {
		return log.NewWithOptions(os.Stderr, log.Options{
			Prefix: fmt.Sprintf("[%.4s]", module),
		})
	} else {
		return log.New(os.Stderr)
	}
}

// NewFileLogger creates a new logger that outputs to a specified file.
func NewFileLogger(fileName string) (*log.Logger, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	lg := log.New(file)
	lg.SetLevel(log.DebugLevel) // Set default level, can be adjusted as needed

	loggers[fileName] = lg
	return lg, nil
}

// flog is a convenience function for logging messages to a specified file logger.
func FDebugf(fileName, format string, args ...interface{}) {
	var logger *log.Logger
	var err error

	logger, exists := loggers[fileName]
	if !exists {
		logger, err = NewFileLogger(fileName)
		if err != nil {
			panic(fmt.Sprintf("Failed to create logger: %v", err))
		}
		loggers[fileName] = logger
	}

	logger.Debugf(format, args...)
}

func SetLogLevel(lvl int) {
	for _, logger := range loggers {
		// fmt.Printf("setting log level to:%v for %v\n ", LoggingLevels[lvl], k)
		logger.SetLevel(LogLvlMap[lvl])

		// Silent mode
		if lvl <= -1 {
			logger.SetOutput(io.Discard)
		} else {
			logger.SetOutput(os.Stderr)
		}
	}
}

// Sets the logging into TUI mode.
func SetTUI(output io.Writer) {
	TUIMode = true
	tuiLogStyles := log.DefaultStyles()
	tuiLogStyles.Levels = logLevelStyles

	tuiLogStyles.Message = logTextStyle
	tuiLogStyles.Value = logTextStyle

	tuiLogStyles.Prefix = logTextFaintStyle
	tuiLogStyles.Key = logTextFaintStyle
	tuiLogStyles.Separator = logTextFaintStyle

	for _, logger := range loggers {
		logger.SetOutput(output)
		// see https://github.com/charmbracelet/log?tab=readme-ov-file#styles
		logger.SetStyles(tuiLogStyles)
		logger.SetColorProfile(termenv.ANSI256)
		logger.SetReportCaller(false)
		logger.SetReportTimestamp(false)
		if logger.GetLevel() < log.InfoLevel {
			logger.SetLevel(log.Level(log.InfoLevel))
		}

	}

}

func init() {
	envDebug := os.Getenv(EnvGosukiDebug)
	if envDebug != "" {
		lvl, err := strconv.Atoi(os.Getenv(EnvGosukiDebug))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s=%v: %v", EnvGosukiDebug, envDebug, err)
		}

		if lvl < -1 {
			lvl = -1
		}
		SetLogLevel(lvl)
	}

	// init global vars
	loggers = make(map[string]*log.Logger)
}
