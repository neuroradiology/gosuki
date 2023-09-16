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
	"fmt"

	"git.blob42.xyz/gosuki/gosuki/pkg/modules"
	"git.blob42.xyz/gosuki/gosuki/pkg/profiles"
	"github.com/urfave/cli/v2"
)



var ProfileCmds = &cli.Command{
	Name: "profile",
	Aliases: []string{"p"},
	Usage: "profile commands",
	Subcommands: []*cli.Command{
		listProfilesCmd,
	},
}


//TODO: only enable commands when modules which implement profiles interfaces
// are available
var listProfilesCmd = &cli.Command{
	Name: "list",
	Usage: "list available profiles",
	Action: func(c *cli.Context) error {

	browsers := modules.GetBrowserModules()
	for _, br := range browsers {

		//Create a browser instance
		brmod, ok := br.ModInfo().New().(modules.BrowserModule)
		if !ok {
			log.Criticalf("module <%s> is not a BrowserModule", br.ModInfo().ID)
		}

		pm, isProfileManager := brmod.(profiles.ProfileManager)
		if !isProfileManager{
			log.Critical("not profile manager")
		}
		if isProfileManager {
			// handle default profile commands

			profs, err := pm.GetProfiles()
			if err != nil {
				return err
			}

			for _, p := range profs {
				fmt.Printf("%-10s \t %s\n", p.Name, pm.GetProfilePath(*p))
			}

			
		}
	}

		return nil
	},
}
