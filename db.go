package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	DB_LOCAL_PATH = "bookmarks.db"
	DB_MEM_PATH   = "file::memory:?cache=shared"
)

var (
	db *sql.DB
)

const (
	CREATE_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		flags integer default 0
	)`
)

// TODO: Use context when making call from request/api
func initInMemoryDb() {
	//sqlite3conn := []*sqlite3.SQLiteConn{}

	var err error

	db, err = sql.Open("sqlite3", DB_MEM_PATH)
	debugPrint("in memory db opened")
	logPanic(err)

	_, err = db.Exec(CREATE_DB_SCHEMA)
	logPanic(err)

}

func initDB() {
	debugPrint("[NotImplemented] initialize local db if not exists or load it")
	// Check if db exists locally

	// If does not exit create new one in memory
	// Then flush to disk

	// If exists locally, load to memory
	initInMemoryDb()
}

func testInMemoryDb() {

	debugPrint("test in memory")
	db, err := sql.Open("sqlite3", DB_MEM_PATH)
	defer db.Close()
	rows, err := db.Query("select URL from bookmarks")
	defer rows.Close()
	logPanic(err)
	var URL string
	for rows.Next() {
		rows.Scan(&URL)
		log.Println(URL)
	}
}

func flushToDisk() {
	/// Flush in memory sqlite db to disk
	/// should happen as often as possible to
	/// avoid losing data
	debugPrint("Flushing to disk")

	conns := []*sqlite3.SQLiteConn{}

	sql.Register("sqlite_with_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conns = append(conns, conn)

				return nil
			},
		})

	memDb, err := sql.Open("sqlite_with_backup", DB_MEM_PATH)
	logPanic(err)
	defer memDb.Close()

	memDb.Ping()

	bkDb, err := sql.Open("sqlite_with_backup", DB_LOCAL_PATH)
	defer bkDb.Close()
	logPanic(err)
	bkDb.Ping()

	bk, err := conns[1].Backup("main", conns[0], "main")
	logPanic(err)

	_, err = bk.Step(-1)
	logPanic(err)

	bk.Finish()

}
