package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

const (
	DB_LOCAL_PATH    = "bookmarks.db"
	DB_MEMCACHE_PATH = "file:memcache?mode=memory&cache=shared"
	DB_BUFFER_PATH   = "file:buffer?mode=memory&cache=shared"
)

var (
	CACHE_DB *DB // Main in memory db, is synced with disc
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

type DB struct {
	name   string
	path   string
	handle *sql.DB
}

func (db *DB) Init() error {

	var err error

	// Create the memory cache db
	db.handle, err = sql.Open("sqlite3", db.path)
	debugPrint("db <%s> opend at at <%s>", db.name, db.path)
	if err != nil {
		return err
	}

	// Populate db schema
	_, err = db.handle.Exec(CREATE_MEM_DB_SCHEMA)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() {
	db.handle.Close()
}

func (db *DB) Count() int {
	var count int

	row := db.handle.QueryRow("select count(*) from bookmarks")
	err := row.Scan(&count)
	logPanic(err)

	return count
}

func (src *DB) SyncTo(dst *DB) error {

	debugPrint("Syncing <%s>(%d) to <%s>(%d)", src.name,
		src.Count(),
		dst.name,
		dst.Count())

	conns := []*sqlite3.SQLiteConn{}

	sql.Register("sqlite_with_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conns = append(conns, conn)

				return nil
			},
		})

	srcDb, err := sql.Open("sqlite_with_backup", src.path)
	defer srcDb.Close()
	if err != nil {
		return err
	}

	srcDb.Ping()

	dstDb, err := sql.Open("sqlite_with_backup", dst.path)
	defer dstDb.Close()
	if err != nil {
		return err
	}
	dstDb.Ping()

	bk, err := conns[1].Backup("main", conns[0], "main")
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

func (db *DB) FlushToDisk() error {

	conns := []*sqlite3.SQLiteConn{}

	sql.Register("sqlite_with_backup",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conns = append(conns, conn)

				return nil
			},
		})

	srcDb, err := sql.Open("sqlite_with_backup", db.path)
	if err != nil {
		return err
	}
	defer srcDb.Close()

	srcDb.Ping()

	bkDb, err := sql.Open("sqlite_with_backup", DB_LOCAL_PATH)
	defer bkDb.Close()
	if err != nil {
		return err
	}
	bkDb.Ping()

	bk, err := conns[1].Backup("main", conns[0], "main")
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

// TODO: Use context when making call from request/api
func initDB() error {
	debugPrint("[NotImplemented] initialize local db if not exists or load it")
	// Check if db exists locally

	// If does not exit create new one in memory
	// Then flush to disk

	// If exists locally, load to memory

	// Initialize memory db with schema
	CACHE_DB = &DB{"memcache",
		DB_MEMCACHE_PATH, nil}

	err := CACHE_DB.Init()
	if err != nil {
		return err
	}

	return nil
}

// TODO: Use context when making call from request/api
//func initMemCacheDb() {
// Init both cache and current memory dbs
// `Cache` is a memory replica of disk db
// `Current` is current working job db
//
//sqlite3conn := []*sqlite3.SQLiteConn{}

//var err error
//memCacheDb := &DB{"memcache",
//DB_MEMCACHE}

//// Create the memory cache db
//memCacheDb.handle, err = sql.Open("sqlite3", memCacheDb.path)
//debugPrint("in memory memCacheDb opened at %s", memCacheDb.path)
//logPanic(err)

//_, err = memCacheDb.Exec(CREATE_MEM_DB_SCHEMA)
//logPanic(err)

// Create the current job db
//currentJobDB, err = sql.Open("sqlite3", DB_CURRENT)
//debugPrint("in memory currentJobDb opened")
//logPanic(err)

//_, err = currentJobDB.Exec(CREATE_MEM_DB_SCHEMA)
//logPanic(err)

//}

func testInMemoryDb(db *DB) {

	debugPrint("test in memory")
	_db, err := sql.Open("sqlite3", db.path)
	defer _db.Close()
	rows, err := _db.Query("select URL from bookmarks")
	defer rows.Close()
	logPanic(err)
	var URL string
	for rows.Next() {
		rows.Scan(&URL)
		log.Println(URL)
	}
}

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
