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

// It is possible to enable debugging for execution time that happens before
// the -debug cli arg is parsed. This is possible using the GOSUKI_DEBUG=X env 
// variable where X is an integer for the debug level
package logging

import (
	"os"

	"strconv"

	glogging "github.com/op/go-logging"
)

var (
	log         = glogging.MustGetLogger("MODE")

    //RELEASE: Change to Release for release mode
	LoggingMode = Debug

)

const EnvGosukiDebug = "GOSUKI_DEBUG"

const Test = -1
const (
	Release = iota
	Info
	Debug
)

func SetMode(lvl int) {
	if lvl > Debug || lvl < -1 {
		log.Warningf("using wrong debug level: %v", lvl)
		return
	}
	LoggingMode = lvl
    setLogLevel(lvl)
}

func initRuntimeMode() {

	envDebug := os.Getenv(EnvGosukiDebug)

	if envDebug != "" {
		mode, err := strconv.Atoi(os.Getenv(EnvGosukiDebug))

		if err != nil {
			log.Errorf("wrong debug level: %v\n%v", envDebug, err)
		}

        SetMode(mode)
	} 

    //TODO: disable debug log when testing
}
