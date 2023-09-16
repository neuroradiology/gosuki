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

package database

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
)

// Print debug Rows results
func DebugPrintRows(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for rows.Next() {
		for i := range cols {
			valuesPtrs[i] = &values[i]
		}
		rows.Scan(valuesPtrs...)

		finalValues := make(map[string]interface{})
		for i, col := range cols {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}

			finalValues[col] = v
		}

		for _, col := range cols {
			fmt.Fprintf(w, "%v\t", finalValues[col])
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
}

// Print debug a single row (does not run rows.next())
func DebugPrintRow(rows *sql.Rows) {
	cols, _ := rows.Columns()
	count := len(cols)
	values := make([]interface{}, count)
	valuesPtrs := make([]interface{}, count)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.Debug)
	for _, col := range cols {
		fmt.Fprintf(w, "%s\t", col)
	}
	fmt.Fprintf(w, "\n")

	for i := 0; i < count; i++ {
		fmt.Fprintf(w, "\t")
	}

	fmt.Fprintf(w, "\n")

	for i, _ := range cols {
		valuesPtrs[i] = &values[i]
	}
	rows.Scan(valuesPtrs...)

	finalValues := make(map[string]interface{})
	for i, col := range cols {
		var v interface{}
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			v = string(b)
		} else {
			v = val
		}

		finalValues[col] = v
	}

	for _, col := range cols {
		fmt.Fprintf(w, "%v\t", finalValues[col])
	}
	fmt.Fprintf(w, "\n")
	w.Flush()
}

func (db *DB) PrintBookmarks() error {

	var url, tags string

	rows, err := db.Handle.Query("select url,tags from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url, &tags)
		if err != nil {
			return err
		}
		log.Debugf("url:%s  tags:%s", url, tags)
	}

	return nil
}
