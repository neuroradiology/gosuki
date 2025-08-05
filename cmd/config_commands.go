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
	"os"

	"github.com/BurntSushi/toml"
	"github.com/blob42/gosuki/pkg/config"
	"github.com/blob42/gosuki/pkg/logging"
	"github.com/kr/pretty"

	"github.com/urfave/cli/v3"
)

var log = logging.GetLogger("cmd")

var cfgPrintCmd = &cli.Command{
	Name:    "gen",
	Aliases: []string{"g"},
	Usage:   "generate a default configuration",
	Action:  printConfig,
}

var cfgDebugCmd = &cli.Command{
	Name:    "debug",
	Aliases: []string{"d"},
	Usage:   "verbose debug of the current config",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		pretty.Print(config.GetAll())
		return nil
	},
}

var ConfigCmds = &cli.Command{
	Name: "config",
	Commands: []*cli.Command{
		cfgPrintCmd,
		cfgDebugCmd,
	},
}

func printConfig(ctx context.Context, cmd *cli.Command) error {
	tomlEncoder := toml.NewEncoder(os.Stdout)
	tomlEncoder.Indent = ""
	outputConf, err := config.MapToOutputConfig(config.GetAll())
	if err != nil {
		return err
	}

	return tomlEncoder.Encode(outputConf)
}
