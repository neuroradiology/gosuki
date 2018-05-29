package main

import (
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// Bookmark type
type Bookmark struct {
	URL      string   `json:"url"`
	Metadata string   `json:"metadata"`
	Tags     []string `json:"tags"`
	Desc     string   `json:"desc"`
	Node     *Node    `json:"-"`
	//flags int
}

// Inserts a bookmarks to the passed DB
// In case of conflict follow the default rules
// which for sqlite is a fail with the error `sqlite3.ErrConstraint`
func (bk *Bookmark) InsertInDB(db *DB) {
	//log.Debugf("Adding bookmark %s", bk.URL)
	_db := db.Handle

	tx, err := _db.Begin()
	logPanic(err)

	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	logError(err)
	defer stmt.Close()

	_, err = stmt.Exec(bk.URL, bk.Metadata, strings.Join(bk.Tags, TagJoinSep), "", 0)
	sqlErrorMsg(err, bk.URL)

	err = tx.Commit()
	logError(err)
}

// Inserts or updates a bookmarks to the passed DB
// In case of a conflict for a UNIQUE URL constraint,
// update the existing bookmark
func (bk *Bookmark) InsertOrUpdateInDB(db *DB) {

	var sqlite3Err sqlite3.Error

	// TODO
	// When updating we should only ADD tags and not replace previous ones

	//log.Debugf("Adding bookmark %s", bk.URL)
	_db := db.Handle

	tx, err := _db.Begin()
	logPanic(err)

	// First try to insert the bookmark (assume it's new)
	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc,
											flags) VALUES (?, ?, ?, ?, ?)`)
	defer stmt.Close()
	sqlErrorMsg(err, bk.URL)

	_, err = stmt.Exec(bk.URL, bk.Metadata, strings.Join(bk.Tags, TagJoinSep), "", 0)

	if err != nil {
		sqlite3Err = err.(sqlite3.Error)
	}

	if err != nil && sqlite3Err.Code != sqlite3.ErrConstraint {
		sqlErrorMsg(err, bk.URL)
	}

	// We will handle ErrConstraint ourselves
	if err != nil && sqlite3Err.Code == sqlite3.ErrConstraint {
		log.Debugf("Updating bookmark %s", bk.URL)
		stmt, err := tx.Prepare(`UPDATE bookmarks SET metadata=?, tags=?
									WHERE url=?`)
		sqlErrorMsg(err, bk.URL)

		_, err = stmt.Exec(bk.Metadata, strings.Join(bk.Tags, TagJoinSep),
			bk.URL)

		sqlErrorMsg(err, bk.URL)
	}

	err = tx.Commit()
	logError(err)

}
