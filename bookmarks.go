package main

import (
	"strings"
)

// Bookmark type
type Bookmark struct {
	Url      string
	Metadata string
	Tags     []string
	Desc     string
	//flags int
}

func (bk *Bookmark) add(db *DB) {
	//log.Debugf("Adding bookmark %s", bk.url)
	_db := db.handle

	tx, err := _db.Begin()
	logPanic(err)

	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	logError(err)
	defer stmt.Close()

	_, err = stmt.Exec(bk.Url, bk.Metadata, strings.Join(bk.Tags, " "), "", 0)
	sqlErrorMsg(err, bk.Url)

	err = tx.Commit()
	logError(err)
}
