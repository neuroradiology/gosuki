package main

import (
	"strings"
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

func (bk *Bookmark) add(db *DB) {
	//log.Debugf("Adding bookmark %s", bk.URL)
	_db := db.handle

	tx, err := _db.Begin()
	logPanic(err)

	// TODO
	// Handle unique constraint errors for when inserting into existing db
	// Should check if err is constraint and return it
	// Or create addAndUpdate function that updates at the same time
	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	logError(err)
	defer stmt.Close()

	_, err = stmt.Exec(bk.URL, bk.Metadata, strings.Join(bk.Tags, TagJoinSep), "", 0)
	sqlErrorMsg(err, bk.URL)

	err = tx.Commit()
	logError(err)
}
