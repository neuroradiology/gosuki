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

package database

// Performs the database schema migration from version 2 to version 3.
// This migration adds synchronization capabilities by:
// 1. Adding a 'version' column to the gskbookmarks table with a default value of 0
// 2. Adding a 'node_id' column to the gskbookmarks table for node identification
// 3. Creating a new 'sync_nodes' table to manage synchronization nodes with:
//   - ordinal (primary key)
//   - node_id (unique constraint)
//   - version (not null)
//
// 4. Populating existing bookmarks with their IDs as version values
func (db *DB) migrateToVersion3() error {
	log.Debug("DB schema: migrating to v3")
	tx, err := db.Handle.Begin()
	if err != nil {
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec("ALTER TABLE gskbookmarks ADD COLUMN version integer default 0;")
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec("ALTER TABLE gskbookmarks ADD COLUMN node_id BLOB;")
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS sync_nodes (
		ordinal INTEGER PRIMARY KEY,
		node_id BLOB NOT NULL UNIQUE,
		version INTEGER NOT NULL
	)`)
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if _, err = tx.Exec(`
		UPDATE gskbookmarks
		SET version = id
		`); err != nil {
		return DBError{DBName: db.Name, Err: err}
	}

	if err := tx.Commit(); err != nil {
		return DBError{DBName: db.Name, Err: err}
	}

	return nil
}
