// Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
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
	"html/template"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/blob42/gosuki"
	db "github.com/blob42/gosuki/internal/database"
)

type searchOpts struct {
	fuzzy bool
}

var FuzzySearchCmd = &cli.Command{
	Name:        "fuzzy",
	Aliases:     []string{"f"},
	Usage:       "fuzzy search anywhere",
	UsageText:   "Uses fuzzy search algorithm on the `URL`, `Title` and `Metadata` fields of the bookmarks database.",
	Description: "",
	ArgsUsage:   "",
	Category:    "",
	Action: func(c *cli.Context) error {
		return searchBookmarks(c, searchOpts{true}, "toto")
	},
}

func formatMark(mark *gosuki.Bookmark, format string) (string, error) {
	outFormat := strings.Clone(format)

	// Comma separated list of tags
	outFormat = strings.ReplaceAll(outFormat, "%T", `{{ join .Tags "," }}`)

	// url
	outFormat = strings.ReplaceAll(outFormat, "%u", `{{.URL}}`)

	// title
	outFormat = strings.ReplaceAll(outFormat, "%t", `{{.Title}}`)

	// description
	outFormat = strings.ReplaceAll(outFormat, "%d", `{{.Desc}}`)

	r := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	outFormat = r.Replace(outFormat)

	return outFormat, nil
}

// Format a bookmark given a fmt.Printf format string
func formatPrint(ctx *cli.Context, marks []*gosuki.Bookmark) error {
	for _, mark := range marks {
		if format := ctx.String("format"); format != "" {
			funcs := template.FuncMap{"join": strings.Join}
			outFormat, err := formatMark(mark, format)
			if err != nil {
				return err
			}

			fmtTmpl, err := template.New("format").Funcs(funcs).Parse(outFormat)
			if err != nil {
				return err
			}

			err = fmtTmpl.Execute(os.Stdout, mark)
			if err != nil {
				return err
			}

			if err != nil {
				return err
			}

		} else {
			fmt.Println(mark.URL)
		}
	}

	return nil
}

func listBookmarks(ctx *cli.Context) error {
	pageParms := db.PaginationParams{
		Page: 1,
		Size: -1,
	}
	result, err := db.ListBookmarks(ctx.Context, &pageParms)
	if err != nil {
		return err
	}

	return formatPrint(ctx, result.Bookmarks)
}

func searchBookmarks(ctx *cli.Context, opts searchOpts, keyword ...string) error {
	pageParms := db.PaginationParams{
		Page: 1,
		Size: -1,
	}
	result, err := db.QueryBookmarks(ctx.Context, keyword[0], opts.fuzzy, &pageParms)
	if err != nil {
		return err
	}
	return formatPrint(ctx, result.Bookmarks)
}
