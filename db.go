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
	DB_BACKUP_HOOK   = "sqlite_with_backup"
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

var _sql3conns []*sqlite3.SQLiteConn // Only used for backup hook

type DB struct {
	name   string
	path   string
	handle *sql.DB
}

func (db *DB) Init() error {

	// TODO: Use context when making call from request/api
	// `CACHE_DB` is a memory replica of disk db
	// `bufferDB` is current working job db

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

	// Check if backup hook has been registered
	backupHookRegistered := false
	for _, val := range sql.Drivers() {
		debugPrint("Checking driver %s", val)
		if val == DB_BACKUP_HOOK {
			backupHookRegistered = true
			break
		}
	}

	if !backupHookRegistered {
		debugPrint("registering driver %s", DB_BACKUP_HOOK)
		sql.Register(DB_BACKUP_HOOK,
			&sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					_sql3conns = append(_sql3conns, conn)
					return nil
				},
			})
	}

	return nil
}

func (db *DB) Close() {
	debugPrint("Closing <%s>", db.name)
	db.handle.Close()
}

func (db *DB) Count() int {
	var count int

	row := db.handle.QueryRow("select count(*) from bookmarks")
	err := row.Scan(&count)
	logPanic(err)

	return count
}

func (db *DB) Print() error {

	var url string

	rows, err := db.handle.Query("select url from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url)
		if err != nil {
			return err
		}
		debugPrint(url)
	}

	return nil
}

func (db *DB) isEmpty() (bool, error) {
	var count int

	row := db.handle.QueryRow("select count(*) from bookmarks")

	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}

	return true, nil
}

func (src *DB) SyncTo(dst *DB) error {

	debugPrint("Syncing <%s>(%d) to <%s>(%d)", src.name,
		src.Count(),
		dst.name,
		dst.Count())

	srcDb, err := sql.Open(DB_BACKUP_HOOK, src.path)
	defer srcDb.Close()
	if err != nil {
		return err
	}

	srcDb.Ping()

	dstDb, err := sql.Open(DB_BACKUP_HOOK, dst.path)
	defer dstDb.Close()
	if err != nil {
		return err
	}
	dstDb.Ping()

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

	debugPrint("Registered Drivers %v", sql.Drivers())
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
