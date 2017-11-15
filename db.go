package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	DB_LOCAL    = "bookmarks.db"
	DB_MEMCACHE = "file:memcache?mode=memory&cache=shared"
	DB_CURRENT  = "file:currentjob?mode=memory&cache=shared&_busy_timeout=5000000"
)

var (
	memCacheDb   *sql.DB
	currentJobDB *sql.DB
)

const (
	// metadata: name or title of resource
	CREATE_LOCAL_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		flags integer default 0
	)`

	CREATE_MEM_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL,
		metadata text default '',
		tags text default '',
		desc text default '',
		flags integer default 0
	)`
)

func isEmptyDb(db *sql.DB) bool {
	var count int

	row := db.QueryRow("select count(*) from bookmarks")

	err := row.Scan(&count)
	logPanic(err)

	if count > 0 {
		return false
	}

	return true
}

func printDBCount(db *sql.DB) {
	var count int

	row := db.QueryRow("select count(*) from bookmarks")
	err := row.Scan(&count)
	logPanic(err)

	debugPrint("%d", count)
}

func printDB(db *sql.DB) {
	var url string

	rows, err := db.Query("select url from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url)
		logPanic(err)
		debugPrint(url)
	}

}

// TODO: Use context when making call from request/api
func initInMemoryDb() {
	// Init both cache and current memory dbs
	// `Cache` is a memory replica of disk db
	// `Current` is current working job db
	//
	//sqlite3conn := []*sqlite3.SQLiteConn{}

	var err error

	// Create the memory cache db
	memCacheDb, err = sql.Open("sqlite3", DB_MEMCACHE)
	debugPrint("in memory memCacheDb opened")
	logPanic(err)

	// Create the current job db
	currentJobDB, err = sql.Open("sqlite3", DB_CURRENT)
	debugPrint("in memory currentJobDb opened")
	logPanic(err)

	_, err = memCacheDb.Exec(CREATE_MEM_DB_SCHEMA)
	logPanic(err)
	_, err = currentJobDB.Exec(CREATE_MEM_DB_SCHEMA)
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
	db, err := sql.Open("sqlite3", DB_MEMCACHE)
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

func syncToDB(src string, dst string) {

	conns := []*sqlite3.SQLiteConn{}

	sql.Register("sqlite_with_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conns = append(conns, conn)

				return nil
			},
		})

	srcDb, err := sql.Open("sqlite_with_backup", src)
	logPanic(err)
	defer srcDb.Close()

	srcDb.Ping()

	dstDb, err := sql.Open("sqlite_with_backup", dst)
	defer dstDb.Close()
	logPanic(err)
	dstDb.Ping()

	bk, err := conns[1].Backup("main", conns[0], "main")
	logPanic(err)

	_, err = bk.Step(-1)
	logPanic(err)

	bk.Finish()
}

func flushToDisk() {
	/// Flush in memory sqlite db to disk
	/// should happen as often as possible to
	/// avoid losing data

	// TODO
	// Memory db and disk db should not be the same
	// Memory db is used to load all urls even duplicates
	debugPrint("Flushing to disk")

	conns := []*sqlite3.SQLiteConn{}

	sql.Register("sqlite_with_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conns = append(conns, conn)

				return nil
			},
		})

	memCacheDb, err := sql.Open("sqlite_with_backup", DB_MEMCACHE)
	logPanic(err)
	defer memCacheDb.Close()

	memCacheDb.Ping()

	bkDb, err := sql.Open("sqlite_with_backup", DB_LOCAL)
	defer bkDb.Close()
	logPanic(err)
	bkDb.Ping()

	bk, err := conns[1].Backup("main", conns[0], "main")
	logPanic(err)

	_, err = bk.Step(-1)
	logPanic(err)

	bk.Finish()

}
