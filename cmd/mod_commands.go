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

package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/blob42/gosuki/pkg/modules"
	"github.com/blob42/gosuki/pkg/profiles"
)

// map cmd Name to *cli.Command
type modCmds map[string]*cli.Command

var (
	// Map registered browser module IDs to their modCmds map
	modCommands = map[string]modCmds{}
)

func RegisterModCommand(modID string, cmd *cli.Command) {
	if cmd == nil {
		log.Fatalf("cannot register nil cmd for <%s>", modID)
	}

	if _, ok := modCommands[modID]; !ok {
		modCommands[modID] = make(modCmds)
	}
	modCommands[modID][cmd.Name] = cmd
}

// return list of registered commands for browser module
func RegisteredModCommands(modID string) modCmds {
	return modCommands[modID]
}

var ModuleCmds = &cli.Command{
	Name:    "modules",
	Aliases: []string{"m"},
	Usage:   "module commands",
	Commands: []*cli.Command{
		listModulesCmd,
	},
}

var listModulesCmd = &cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list available browsers and modules",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s\n", "available browsers and modules:")
		mods := modules.GetModules()
		for _, mod := range mods {
			browser, isBrowser := mod.ModInfo().New().(modules.BrowserModule)
			if isBrowser {
				fmt.Printf(" %s %-15s \tbrowser\n", green(""), mod.ModInfo().ID)
				if pm, ok := browser.(profiles.ProfileManager); ok {
					flavours := pm.ListFlavours()
					if len(flavours) > 1 {
						fmt.Printf("  └─ flavours:\t")
					}
					for _, f := range flavours {
						if f.Flavour == string(mod.ModInfo().ID) {
							continue
						}
						fmt.Printf("%s ", color.New(color.Underline).SprintFunc()(f.Flavour))
					}
					println()
				}
			} else {
				fmt.Printf(" %s %-15s \tmodule\n", green(""), mod.ModInfo().ID)
			}
		}
		return nil
	},
}
