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

// TODO: add cli options to set/get options
// TODO: move browser module commands to their own module packag
package firefox

import (
	"github.com/blob42/gosuki/cmd"
	"github.com/blob42/gosuki/pkg/browsers/mozilla"
	"github.com/blob42/gosuki/pkg/logging"

	"github.com/urfave/cli/v2"
)

var fflog = logging.GetLogger("FF")

var (
	ffUnlockVFSCmd = cli.Command{
		Name:    "unlock",
		Usage:   "Remove VFS lock from places.sqlite",
		Aliases: []string{"u"},
		Action:  ffUnlockVFS,
	}

	ffCheckVFSCmd = cli.Command{
		Name:    "check",
		Aliases: []string{"c"},
		Action:  ffCheckVFS,
	}

	ffVFSCommands = cli.Command{
		Name:  "vfs",
		Usage: "VFS locking commands",
		Subcommands: []*cli.Command{
			&ffUnlockVFSCmd,
			&ffCheckVFSCmd,
		},
	}
)

var FirefoxCmds = &cli.Command{
	Name:    "firefox",
	Aliases: []string{"ff"},
	Usage:   "firefox related commands",
	Subcommands: []*cli.Command{
		&ffVFSCommands,
	},
}

func init() {
	cmd.RegisterModCommand(BrowserName, FirefoxCmds)
}

// TODO: #54 define interface for modules to handle and list profiles

func ffCheckVFS(_ *cli.Context) error {
	err := mozilla.CheckVFSLock("path to profile")
	if err != nil {
		return err
	}

	return nil
}

func ffUnlockVFS(_ *cli.Context) error {
	err := mozilla.UnlockPlaces("path to profile")
	if err != nil {
		return err
	}

	return nil
}
