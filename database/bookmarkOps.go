package database

import (
	"git.sp4ke.xyz/sp4ke/gomark/bookmarks"
	"strings"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const TagJoinSep = ","

type Bookmark = bookmarks.Bookmark

// Inserts or updates a bookmarks to the passed DB
// In case of a conflict for a UNIQUE URL constraint,
// update the existing bookmark
func (db *DB) InsertOrUpdateBookmark(bk *Bookmark) {
	var sqlite3Err sqlite3.Error
	var scannedTags string

	_db := db.Handle

	// Prepare statement that does a pure insert only
	tryInsertBk, err := _db.Prepare(
		`INSERT INTO
			bookmarks(URL, metadata, tags, desc, flags)
			VALUES (?, ?, ?, ?, ?)`,
	)
	defer tryInsertBk.Close()
	if err != nil {
		log.Errorf("%s: %s", err, bk.URL)
	}

	// Prepare statement that updates an existing bookmark in db
	updateBk, err := _db.Prepare(
		`UPDATE bookmarks SET metadata=?, tags=?, modified=strftime('%s')
		WHERE url=?`,
	)
	defer updateBk.Close()
	if err != nil {
		log.Errorf("%s: %s", err, bk.URL)
	}

	// Stmt to fetch existing bookmark and tags in db
	getTags, err := _db.Prepare(`SELECT tags FROM bookmarks WHERE url=? LIMIT 1`)
	defer getTags.Close()
	if err != nil {
		log.Errorf("%s: %s", err, bk.URL)
	}

	// Begin transaction
	tx, err := _db.Begin()
	if err != nil {
		log.Error(err)
	}

	// First try to insert the bookmark (assume it's new)
    tagList := strings.Join(bk.Tags, TagJoinSep)
	_, err = tx.Stmt(tryInsertBk).Exec(
		bk.URL,
		bk.Metadata,
		tagList,
		"", 0,
	)

	if err != nil {
		sqlite3Err = err.(sqlite3.Error)
	}

	if err != nil && sqlite3Err.Code != sqlite3.ErrConstraint {
		log.Errorf("%s: %s", err, bk.URL)
	}
	// We will handle ErrConstraint ourselves

	// ErrConstraint means the bookmark (url) already exists in table,
	// we need to update it instead.
	if err != nil && sqlite3Err.Code == sqlite3.ErrConstraint {
		log.Debugf("Updating bookmark %s", bk.URL)

		// First get existing tags for this bookmark if any ?
		res := tx.Stmt(getTags).QueryRow(
			bk.URL,
		)
		res.Scan(&scannedTags)
		cacheTags := strings.Split(scannedTags, TagJoinSep)

		// If tags are different, merge current bookmark tags and existing tags
		// Put them in a map first to remove duplicates
		tagMap := make(map[string]bool)
		for _, v := range cacheTags {
			tagMap[v] = true
		}
		for _, v := range bk.Tags {
			tagMap[v] = true
		}

		var newTags []string // merged tags

		// Merge in a single slice
		for k, _ := range tagMap {
			newTags = append(newTags, k)
		}

		_, err = tx.Stmt(updateBk).Exec(
			bk.Metadata,
			strings.Join(newTags, TagJoinSep), // Join tags with a `|`
			bk.URL,
		)

		if err != nil {
			log.Errorf("%s: %s", err, bk.URL)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}

}

// Inserts a bookmarks to the passed DB
// In case of conflict follow the default rules
// which for sqlite is a fail with the error `sqlite3.ErrConstraint`
func (db *DB) InsertBookmark(bk *Bookmark) {
	//log.Debugf("Adding bookmark %s", bk.URL)
	_db := db.Handle
	tx, err := _db.Begin()
	if err != nil {
		log.Error(err)
	}

	stmt, err := tx.Prepare(`INSERT INTO bookmarks(URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Error(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(bk.URL, bk.Metadata, strings.Join(bk.Tags, TagJoinSep), "", 0)
	if err != nil {
		log.Errorf("%s: %s", err, bk.URL)
	}

	err = tx.Commit()
	if err != nil {
		log.Error(err)
	}
}
