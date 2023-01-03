package database

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// For ever row in `src` try to insert it into `dst`.
// If if fails then try to update it. It means `src` is synced to `dst`
func (src *DB) SyncTo(dst *DB) {
	var sqlite3Err sqlite3.Error
	var existingUrls []*SBookmark

	log.Debugf("syncing <%s> to <%s>", src.Name, dst.Name)

	getSourceTable, err := src.Handle.Prepare(`SELECT * FROM bookmarks`)
	defer func() {
		err = getSourceTable.Close()
		if err != nil {
			log.Critical(err)
		}
	}()
	if err != nil {
		log.Error(err)
	}

	getDstTags, err := dst.Handle.Prepare(
		`SELECT tags FROM bookmarks WHERE url=? LIMIT 1`,
	)

	defer func() {
		err := getDstTags.Close()

		if err != nil {
			log.Critical(err)
		}
	}()

	if err != nil {
		log.Error(err)
	}

	tryInsertDstRow, err := dst.Handle.Prepare(
		`INSERT INTO
		bookmarks(url, metadata, tags, desc, flags)
		VALUES (?, ?, ?, ?, ?)`,
	)
	defer func() {
		err := tryInsertDstRow.Close()
		if err != nil {
			log.Critical(err)
		}
	}()

	if err != nil {
		log.Error(err)
	}

	updateDstRow, err := dst.Handle.Prepare(
		`UPDATE bookmarks
		SET (metadata, tags, desc, modified, flags) = (?,?,?,strftime('%s'),?)
		WHERE url=?
		`,
	)

    defer func(){
        err := updateDstRow.Close()
        if err != nil {
            log.Critical()
        }
    }()
    
	if err != nil {
		log.Error(err)
	}

	srcTable, err := getSourceTable.Query()
	if err != nil {
		log.Error(err)
	}

	// Lock destination db
	dstTx, err := dst.Handle.Begin()
	if err != nil {
		log.Error(err)
	}

	// Start syncing all entries from source table
	for srcTable.Next() {

		// Fetch on row
		scan, err := ScanBookmarkRow(srcTable)
		if err != nil {
			log.Error(err)
		}

		// Try to insert to row in dst table
		_, err = dstTx.Stmt(tryInsertDstRow).Exec(
			scan.Url,
			scan.metadata,
			scan.tags,
			scan.desc,
			scan.flags,
		)

		if err != nil {
			sqlite3Err = err.(sqlite3.Error)
		}

		if err != nil && sqlite3Err.Code != sqlite3.ErrConstraint {
			log.Error(err)
		}

		// Record already exists in dst table, we need to use update
		// instead.
		if err != nil && sqlite3Err.Code == sqlite3.ErrConstraint {
			existingUrls = append(existingUrls, scan)
		}
	}

	err = dstTx.Commit()
	if err != nil {
		log.Error(err)
	}

	// Start a new transaction to update the existing urls
	dstTx, err = dst.Handle.Begin() // Lock dst db
	if err != nil {
		log.Error(err)
	}

	// Traverse existing urls and try an update this time
	for _, scan := range existingUrls {
		var tags string

		//log.Debugf("updating existing %s", scan.Url)

		row := getDstTags.QueryRow(
			scan.Url,
		)
		row.Scan(&tags)

		//log.Debugf("src tags: %v", scan.tags)
		//log.Debugf("dst tags: %v", dstTags)
		srcTags := strings.Split(scan.tags, TagJoinSep)
		dstTags := strings.Split(tags, TagJoinSep)
		tagMap := make(map[string]bool)
		for _, v := range srcTags {
			tagMap[v] = true
		}
		for _, v := range dstTags {
			tagMap[v] = true
		}

		var newTags []string //merged tags
		for k := range tagMap {
			newTags = append(newTags, k)
		}

		_, err = dstTx.Stmt(updateDstRow).Exec(
			scan.metadata,
			strings.Join(newTags, TagJoinSep),
			scan.desc,
			0, //flags
			scan.Url,
		)

		if err != nil {
			log.Errorf("%s: %s", err, scan.Url)
		}

	}

	err = dstTx.Commit()
	if err != nil {
		log.Error(err)
	}

	// If we are syncing to memcache, sync cache to disk
	if dst.Name == CacheName {
		err = dst.SyncToDisk(GetDBFullPath())
		if err != nil {
			log.Error(err)
		}
	}
}

func (src *DB) SyncToDisk(dbpath string) error {
	log.Debugf("Syncing <%s> to <%s>", src.Name, dbpath)

	//log.Debugf("[flush] openeing <%s>", src.path)
	srcDb, err := sqlx.Open(DriverBackupMode, src.Path)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//log.Debugf("[flush] opening <%s>", DB_FILENAME)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	bkDb, err := sqlx.Open(DriverBackupMode, dbUri)
	defer flushSqliteCon(bkDb)
	if err != nil {
		return err
	}

	err = bkDb.Ping()
	if err != nil {
		return err
	}

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		return err
	}

	_, err = bk.Step(-1)
	if err != nil {
		return err
	}

	bk.Finish()

	return nil
}

func (dst *DB) SyncFromDisk(dbpath string) error {

	log.Debugf("Syncing <%s> to <%s>", dbpath, dst.Name)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	srcDb, err := sqlx.Open(DriverBackupMode, dbUri)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//log.Debugf("[flush] opening <%s>", DB_FILENAME)
	bkDb, err := sqlx.Open(DriverBackupMode, dst.Path)
	defer flushSqliteCon(bkDb)
	if err != nil {
		return err
	}
	bkDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		return err
	}

	_, err = bk.Step(-1)
	if err != nil {
		return err
	}

	bk.Finish()

	return nil
}

// Copy from src DB to dst DB
// Source DB os overwritten
func (src *DB) CopyTo(dst *DB) {

	log.Debugf("Copying <%s> to <%s>", src.Name, dst.Name)

	srcDb, err := sqlx.Open(DriverBackupMode, src.Path)
	defer func() {
		srcDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	if err != nil {
		log.Error(err)
	}

	srcDb.Ping()

	dstDb, err := sqlx.Open(DriverBackupMode, dst.Path)
	defer func() {
		dstDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	if err != nil {
		log.Error(err)
	}
	dstDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	if err != nil {
		log.Error(err)
	}

	_, err = bk.Step(-1)
	if err != nil {
		log.Error(err)
	}

	bk.Finish()
}
