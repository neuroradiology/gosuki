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

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	db "github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/internal/utils"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli/v3"
)

var BukuCmds = &cli.Command{
	Name:    "buku",
	Aliases: []string{"bu"},
	Usage:   "commands for interacting with buku ecosystem",
	Commands: []*cli.Command{
		importBukuDBCmd,
	},
}

var importBukuDBCmd = &cli.Command{
	Name:    "import",
	Aliases: []string{"imp"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "path",
			Usage:       "path to buku DB",
			TakesFile:   true,
			Value:       "~/.local/share/buku/bookmarks.db",
			DefaultText: "~/.local/share/buku/bookmarks.db",
		},
	},
	Usage: `Upgrades a buku DB into a gosuki DB.
	
The upgraded Gosuki DB will stay compatible with buku. A backup of the original
buku database will be made and stored in the same buku data directory.
`,
	Action: importBukuDB,
}

// Imports a buku database to gosuki database
func importBukuDB(ctx context.Context, cmd *cli.Command) error {
	path := cmd.String("path")
	if path == "" {
		return fmt.Errorf("no path provided")
	}

	expandedPath, err := utils.ExpandPath(path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	if _, err = os.Stat(expandedPath); os.IsNotExist(err) {
		return fmt.Errorf("buku DB does not exist at %s", expandedPath)
	}
	bukuDB, err := sqlx.Open("sqlite3", expandedPath)
	if err != nil {
		return fmt.Errorf("failed to connect to Buku DB: %w", err)
	}
	defer bukuDB.Close()

	var bukuBookmarks []struct {
		ID       int64  `db:"id"`
		URL      string `db:"URL"`
		Metadata string `db:"metadata"`
		Tags     string `db:"tags"`
		Desc     string `db:"desc"`
	}
	if err = bukuDB.Select(&bukuBookmarks, "SELECT id, URL, metadata, tags, desc FROM bookmarks"); err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	db.Init()

	DB := db.DiskDB
	defer db.DiskDB.Close()

	fmt.Printf(`Importing bookmarks from buku → gosuki
buku DB path	: %s
gosuki DB path	: %s
...
`, path, db.GetDBFullPath())

	var bkCount int
	for _, b := range bukuBookmarks {
		bk := &db.Bookmark{
			URL:    b.URL,
			Title:  b.Metadata,
			Tags:   strings.Split(b.Tags, ","),
			Desc:   b.Desc,
			Module: "buku",
		}

		if err := DB.UpsertBookmark(bk); err != nil {
			return fmt.Errorf("failed to insert bookmark: %w", err)
		} else {
			bkCount++
		}
	}

	fmt.Printf("Successfully imported %d Buku bookmarks into Gosuki DB\n", bkCount)
	return nil
}
