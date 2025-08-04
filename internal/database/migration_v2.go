// Create a new migration function to from v2 to v3
// This is the new database schema:
// QCreateSchema = `
//
//	   CREATE TABLE IF NOT EXISTS gskbookmarks (
//		id INTEGER PRIMARY KEY,
//		URL TEXT NOT NULL UNIQUE,
//		metadata TEXT default '',
//		tags TEXT default '',
//		desc TEXT default '',
//		modified INTEGER DEFAULT (strftime('%s')),
//		flags INTEGER DEFAULT 0,
//		module TEXT DEFAULT '' ,
//		xhsum TEXT DEFAULT '',
//		version integer default 0, -- new column
//		node_id integer default 0 -- new column
//
// );
//
// CREATE TABLE IF NOT EXISTS sync_nodes (
//
//	ordinal INTEGER PRIMARY KEY,
//	node_id BLOB NOT NULL UNIQUE,
//	version INTEGER NOT NULL
//
// )
// `
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
		return DBError{DBName: db.Name, Err: err}
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
