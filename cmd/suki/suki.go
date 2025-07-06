// Copyright (c) 2024-2025-2025-2025-2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with gosuki.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/blob42/gosuki/internal/database"
	db "github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/pkg/build"
	"github.com/blob42/gosuki/pkg/config"
	"github.com/blob42/gosuki/pkg/logging"
)

func main() {
	app := cli.NewApp()
	app.Version = build.Version()

	app.Name = "suki"
	app.Description = "TODO: summary gosuki description"
	app.Usage = "swiss-knife bookmark manager - cli"
	app.UsageText = "suki [OPTIONS] [KEYWORD [KEYWORD...]] "
	app.CustomAppHelpTemplate = AppHelpTemplate

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:        "debug",
			Category:    "",
			Aliases:     []string{"d"},
			DefaultText: "0",
			Usage:       "set debug level. (`0`-3)",
			EnvVars:     []string{logging.EnvGosukiDebug},
			Action: func(_ *cli.Context, val int) error {
				logging.SetLogLevel(val)
				return nil
			},
		},
		&cli.StringFlag{
			Name:     "format",
			Category: "",
			Usage:    "Format output using a custom template",
			Aliases:  []string{"f"},
		},
	}

	app.Before = func(c *cli.Context) error {
		config.Init()
		db.RegisterSqliteHooks()
		err := db.InitDiskConn(db.GetDBPath())
		if _, isDBErr := err.(database.DBError); isDBErr {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr, "Did you run `gosuki start` at least once ?")
			os.Exit(10)
		}

		return err
	}

	app.Commands = []*cli.Command{
		FuzzySearchCmd,
	}

	app.Action = func(c *cli.Context) error {
		// if no argument passed, list all bookmarks
		//TODO: no arg => interactive cli
		if c.Args().Len() == 0 {
			return listBookmarks(c)
		}

		// use ~ as fuzzy character
		firstKw := c.Args().First()
		opts := searchOpts{}

		if firstKw[0] == '~' {
			opts.fuzzy = true
			firstKw = firstKw[1:]
		}

		return searchBookmarks(c, opts, firstKw)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func init() {
	config.RegisterModuleOpt("database", "db-path", db.DefaultDBPath)
}
