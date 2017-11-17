package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// Global cache database
var (
	CACHE_DB *DB // Main in memory db, is synced with disc
)

const (
	DB_FILENAME      = "gomarks.db"
	DB_MEMCACHE_PATH = "file:memcache?mode=memory&cache=shared"
	DB_BUFFER_PATH   = "file:buffer?mode=memory&cache=shared"
	DB_BACKUP_HOOK   = "sqlite_with_backup"
)

// DB SCHEMAS
const (
	// metadata: name or title of resource
	CREATE_LOCAL_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		modified integer default ?,
		flags integer default 0
	)`

	CREATE_MEM_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL,
		metadata text default '',
		tags text default '',
		desc text default '',
		modified integer default (strftime('%s')),
		flags integer default 0
	)`
)

var _sql3conns []*sqlite3.SQLiteConn // Only used for backup hook
var BACKUPHOOK_REGISTERED bool

type DB struct {
	name         string
	path         string
	handle       *sql.DB
	backupHookOn bool
}

func (db DB) New(name string, path string) *DB {
	return &DB{name, path, nil, false}
}

func (db *DB) Error() string {
	errMsg := fmt.Sprintf("[error][db] name <%s>", db.name)
	return errMsg
}

// Initialize a sqlite database
func (db *DB) Init() {

	// TODO: Use context when making call from request/api
	// `CACHE_DB` is a memory replica of disk db
	// `bufferDB` is current working job db

	var err error

	if db.handle != nil {
		//dbError = DBError()
		logError(db, "already initialized")
		return
	}

	// Create the memory cache db
	db.handle, err = sql.Open("sqlite3", db.path)
	debugPrint("db <%s> opend at at <%s>", db.name, db.path)
	logPanic(err)

	// Populate db schema

	tx, err := db.handle.Begin()
	logPanic(err)

	stmt, err := tx.Prepare(CREATE_MEM_DB_SCHEMA)
	logPanic(err)

	_, err = stmt.Exec()
	logPanic(err)

	err = tx.Commit()
	logPanic(err)

	// Check if backup hook has been registered
	//for _, val := range sql.Drivers() {
	//debugPrint("Checking driver %s", val)
	//if val == DB_BACKUP_HOOK {
	//db.backupHookOn == true
	//break
	//}
	//}

	if !BACKUPHOOK_REGISTERED {
		debugPrint("backup_hook: registering driver %s", DB_BACKUP_HOOK)
		// Register the hook
		sql.Register(DB_BACKUP_HOOK,
			&sqlite3.SQLiteDriver{
				ConnectHook: func(conn *sqlite3.SQLiteConn) error {
					//debugPrint("[HOOK] registering new connection")
					_sql3conns = append(_sql3conns, conn)
					//debugPrint("%v", _sql3conns)
					return nil
				},
			})
		BACKUPHOOK_REGISTERED = true
	}

	debugPrint("<%s> initialized", db.path)
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

	rows, err := db.handle.Query("select url, modified from bookmarks")

	for rows.Next() {
		err = rows.Scan(&url)
		if err != nil {
			return err
		}
		debugPrint("%s", url)
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

func (src *DB) SyncTo(dst *DB) {

	debugPrint("Syncing <%s>(%d) to <%s>(%d)", src.name,
		src.Count(),
		dst.name,
		dst.Count())

	srcDb, err := sql.Open(DB_BACKUP_HOOK, src.path)
	defer func() {
		srcDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	logPanic(err)

	srcDb.Ping()

	dstDb, err := sql.Open(DB_BACKUP_HOOK, dst.path)
	defer func() {
		dstDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	logPanic(err)
	dstDb.Ping()

	bk, err := _sql3conns[1].Backup("main", _sql3conns[0], "main")
	logPanic(err)

	_, err = bk.Step(-1)
	logPanic(err)

	bk.Finish()
}

func (src *DB) FlushToDisk() error {

	if !BACKUPHOOK_REGISTERED {
		errMsg := fmt.Sprintf("%s, %s", src.path, "db backup hook is not initialized")
		return errors.New(errMsg)
	}

	//debugPrint("[flush] openeing <%s>", src.path)
	srcDb, err := sql.Open(DB_BACKUP_HOOK, src.path)
	defer func() {
		srcDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
	if err != nil {
		return err
	}
	srcDb.Ping()

	//debugPrint("[flush] opening <%s>", DB_FILENAME)
	bkDb, err := sql.Open(DB_BACKUP_HOOK, DB_FILENAME)
	defer func() {
		bkDb.Close()
		_sql3conns = _sql3conns[:len(_sql3conns)-1]
	}()
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

// TODO: Use context when making call from request/api
// TODO: Initialize local db
func initDB() {
	debugPrint("[NotImplemented] initialize local db if not exists or load it")
	//debugPrint("Registered Drivers %v", sql.Drivers())

	// If does not exit create new one in memory
	// Then flush to disk

	// If exists locally, load to memory

	// Initialize memory db with schema
	CACHE_DB = DB{}.New("memcache", DB_MEMCACHE_PATH)
	CACHE_DB.Init()
	debugPrint("Registered Drivers %v", sql.Drivers())

	// Check and initialize local db as last step
	// after loading the different browser bookmarks to cache

	dbdir := getDefaultDBPath()
	err := checkWriteable(dbdir)
	logPanic(err)

	dbpath := filepath.Join(dbdir, DB_FILENAME)

	// If local db exists load it to CACHE_DB
	var exists bool
	if exists, err = checkFileExists(dbpath); exists {
		logPanic(err)
		debugPrint("[NOT IMPLEMENTED] preload existing local db")
	} else {
		logPanic(err)
		// Else initialize it
		initLocalDB(CACHE_DB, dbpath)
	}

}

func initLocalDB(db *DB, dbpath string) {

	debugPrint("Initializing local db at '%s'", dbpath)
	debugPrint("%s flushing to disk", db.name)
	err := db.FlushToDisk()
	logPanic(err)

	// DEBUG
	//debugPrint("flushing again")
	//time.Sleep(2 * time.Second)
	//_ = db.FlushToDisk()

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
