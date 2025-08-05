//
//  Copyright (c) 2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package logging

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/urfave/cli/v3"
)

var DebugFlag = &cli.StringFlag{
	Name:        "debug",
	Usage:       debugHelp,
	DefaultText: "warn",
	Sources:     cli.EnvVars(EnvGosukiDebug),
	Action: func(_ context.Context, _ *cli.Command, val string) error {
		if SilentMode {
			SetLevel(Silent)
		}
		return ParseDebugLevels(val)
	},
}

// errors
var (
	ErrUnknownLevel  = errors.New("unknown debug level")
	ErrHelpQuit      = errors.New("help quit")
	ErrParseSubLevel = errors.New("cannot parse unit level")
)

func parseLevel(lvl string) (string, error) {
	if slices.Contains(allLevels, lvl) {
		return allLevels[slices.Index(allLevels, lvl)], nil
	} else {
		return "", ErrUnknownLevel
	}
}

func parseUnitLvl(sl string) error {
	tokens := strings.Split(sl, "=")
	// fmt.Printf("%#v\n", tokens)

	if len(tokens) != 2 {
		return ErrParseSubLevel
	}
	unit, lvl := tokens[0], tokens[1]

	if !slices.Contains(allLevels, lvl) {
		return fmt.Errorf("%w %s", ErrUnknownLevel, lvl)
	}
	loggerLevels[unit] = levels[lvl]
	SetUnitLevel(unit, levels[lvl])

	return nil
}

var debugHelp = `Logging level for all units {trace, debug, info, warn, error, fatal, none}
	You may also specify <global-level>,<unit>=<level>,<unit2>=<level>,...
	Use 'debug=list' to list available units`

func ParseDebugLevels(val string) error {
	var err error
	var global string
	args := strings.Split(val, ",")

	if args[0] == "list" {
		fmt.Printf("available levels: [%s]\n", strings.Join(allLevels, ","))
		fmt.Printf("available units: [%s]\n", strings.Join(listLoggers(), ","))
		return ErrHelpQuit
	}

	// parse global lvl
	if global, err = parseLevel(args[0]); err != nil {
		return fmt.Errorf("%w `%s'", err, args[0])
	} else {
		globalLevel = levels[global]
		SetLevel(levels[global])
	}

	// subsystem levels
	if len(args) > 1 {
		for _, arg := range args[1:] {
			if err = parseUnitLvl(arg); err != nil {
				return fmt.Errorf("%w `%s'", err, arg)
			}
		}
	}

	// fmt.Printf("%#v\n", args)
	// fmt.Printf("global: %s\n", global)
	// fmt.Printf("sub levels: %v\n", loggerLevels)
	// os.Exit(42)
	return nil
}
