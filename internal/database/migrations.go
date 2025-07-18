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

// - adds the `xhsum` column to `gskbookmarks` and calculates the xhsum,
// - restores `id` primary key column on gskbookmarks
// - `URL` has unique constraint instead of being pk
func (db *DB) migrateToVersion2() error {
	log.Debug("DB schema: migrating to v2")
	tx, err := db.Handle.Begin()
	if err != nil {
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec(`
		DROP VIEW IF EXISTS bookmarks;
		DROP TRIGGER IF EXISTS bookmarks_insert;
		DROP TRIGGER IF EXISTS bookmarks_update;
		CREATE TABLE temp_gskbookmarks (
			id INTEGER PRIMARY KEY,
			URL TEXT NOT NULL UNIQUE,
			metadata TEXT DEFAULT '',
			tags TEXT DEFAULT '',
			desc TEXT DEFAULT '',
			modified INTEGER DEFAULT (strftime('%s')),
			flags INTEGER DEFAULT 0,
			module TEXT DEFAULT '',
			xhsum TEXT DEFAULT ''
		    );
		`)
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec(`
        INSERT INTO temp_gskbookmarks (URL, metadata, tags, desc, modified, flags, module, xhsum)
        SELECT URL, metadata, tags, desc, modified, flags, module, ''
        FROM gskbookmarks
    `)
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec(`
        UPDATE temp_gskbookmarks
        SET xhsum = xhash(printf('%s+%s+%s+%s', URL, metadata, tags, desc))
    `)
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec("DROP TABLE gskbookmarks")
	if err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	_, err = tx.Exec("ALTER TABLE temp_gskbookmarks RENAME TO gskbookmarks")
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec(QCreateView); err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if _, err = tx.Exec(QCreateInsertTrigger); err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if _, err = tx.Exec(QCreateUpdateTrigger); err != nil {
		tx.Rollback()
		return DBError{DBName: db.Name, Err: err}
	}

	if err = tx.Commit(); err != nil {
		return DBError{DBName: db.Name, Err: err}

	}

	return nil
}
