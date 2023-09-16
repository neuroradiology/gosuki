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

// Main command line entry point for gosuki
package main

import (
	"os"

	"git.blob42.xyz/gosuki/gosuki/internal/config"
	"git.blob42.xyz/gosuki/gosuki/internal/logging"
	"git.blob42.xyz/gosuki/gosuki/pkg/modules"
	"git.blob42.xyz/gosuki/gosuki/internal/utils"

	"git.blob42.xyz/gosuki/gosuki/cmd"

	"github.com/urfave/cli/v2"

	// Load firefox browser modules
	_ "git.blob42.xyz/gosuki/gosuki/browsers/firefox"

	// Load chrome browser module
	_ "git.blob42.xyz/gosuki/gosuki/browsers/chrome"
)

var log = logging.GetLogger("")


func main() {

	app := cli.NewApp()


	app.Name = "gosuki"
	app.Version = utils.Version()

	flags := []cli.Flag{

		&cli.StringFlag{
			Name:  "config-file",
			Value: config.ConfigFile,
			Usage: "TOML config `FILE` path",
		},

        &cli.IntFlag{
        	Name:        "debug",
        	Aliases:     []string{"d"},
        	EnvVars:     []string{logging.EnvGosukiDebug},
            Action: func (c *cli.Context, val int) error {
                logging.SetMode(val)
                return nil
            },

        },
	}

	flags = append(flags, config.SetupGlobalFlags()...)
	app.Flags = append(app.Flags, flags...)

	app.Before = func(c *cli.Context) error {


		// get all registered browser modules
		modules := modules.GetModules()
		for _, mod := range modules {

			// Run module's hooks that should run before context is ready
			// for example setup flags management
			modinfo := mod.ModInfo()
			hook := cmd.BeforeHook(string(modinfo.ID))
			if hook != nil {
				if err := cmd.BeforeHook(string(modinfo.ID))(c); err != nil {
					return err
				}
			}
		}

		// Execute config hooks
		//TODO: better doc for what are Conf hooks ???
		config.RunConfHooks(c)

		initConfig()

		return nil
	}


	// Browser modules can register commands through cmd.RegisterModCommand.
	// registered commands will be appended here
	app.Commands = []*cli.Command{
		// main entry point
		startDaemonCmd,
		cmd.ConfigCmds,
		cmd.ProfileCmds,
		cmd.ModuleCmds,
	}

	// Add global flags from registered modules
	// we use GetModules to handle all types of modules
	modules := modules.GetModules()
	log.Debugf("loading %d modules", len(modules))
	for _, mod := range modules {
		modID := string(mod.ModInfo().ID)
		log.Debugf("loading module <%s>", modID)

		// for each registered module, register own flag management
		modFlags := cmd.GlobalFlags(modID)
		if len(modFlags) != 0 {
			app.Flags = append(app.Flags, modFlags...)
		}

		// Add all browser module registered commands
		cmds := cmd.RegisteredModCommands(modID)
		for _, cmd := range cmds {
			app.Commands = append(app.Commands, cmd)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func init() {

	//TODO: watch all profiles (handled at browser level for now)
	// config.RegisterGlobalOption("all-profiles", false)
}

