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

package cmd

import (
	"github.com/urfave/cli/v2"
)

var modFlags = map[string][]cli.Flag{}

// RegGlobalModFlag registers global flags to pass on to the browser module
func RegGlobalModFlag(modID string, flag cli.Flag) {
	if flag == nil {
		log.Fatal("registering nil flag")
	}

	log.Debugf("<%s> registering global flag: %s",
		modID,
		flag.Names())
	if _, ok := modFlags[modID]; !ok {
		modFlags[modID] = []cli.Flag{flag}
	} else {
		modFlags[modID] = append(modFlags[modID], flag)
	}
}

// GlobalFlags returns the registered global flags for a given registered module
func GlobalFlags(modID string) []cli.Flag {
	return modFlags[modID]
}
