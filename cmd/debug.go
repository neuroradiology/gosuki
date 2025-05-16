//
//  Copyright (c) 2024 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
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
	"github.com/kr/pretty"
	"github.com/urfave/cli/v2"

	"github.com/blob42/gosuki/pkg/marktab"
)

var DebugCmd = &cli.Command{
	Name:  "debug",
	Usage: "debug",
	Action: func(_ *cli.Context) error {
		var err error
		// db.Init()
		// res := ""
		// IMP: Need DriverTest to be used for opening DiskDB
		// row := db.DiskDB.Handle.QueryRowx(`SELECT fuzzy('w3r', 'W3M Rocks')`)
		// err = row.Scan(&res)
		// if err != nil {
		// 	return err
		// }
		// fmt.Println(res)

		// marktabs
		mt := marktab.MarkTab{}
		err = mt.LoadMarktabs()
		if err != nil {
			return err
		}

		pretty.Print(mt)
		return err
	},
}
