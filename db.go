package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	sqlite3 "github.com/mattn/go-sqlite3"
)

// Global cache database
var (
	cacheDB              *DB                   // Main in memory db, is synced with disc
	_sql3conns           []*sqlite3.SQLiteConn // Only used for backup hook
	backupHookRegistered bool                  // set to true once the backup hook is registered
)

const (
	DB_FILENAME      = "gomarks.db"
	DB_MEMCACHE_PATH = "file:memcache?mode=memory&cache=shared"
	DB_BUFFER_PATH   = "file:buffer?mode=memory&cache=shared"
	DB_BACKUP_HOOK   = "sqlite_with_backup"
)

//  Database schemas used for the creation of new databases
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
		URL text NOT NULL UNIQUE,
		metadata text default '',
		tags text default '',
		desc text default '',
		modified integer default (strftime('%s')),
		flags integer default 0
	)`
)

// DB encapsulates an sql.DB
// all itneractions with memory/buffer and disk databases
// is done through the DB object
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
	// `cacheDB` is a memory replica of disk db
	// `bufferDB` is current working job db

	var err error

	if db.handle != nil {
		logErrorMsg(db, "already initialized")
		return
	}

	// Create the memory cache db
	db.handle, err = sql.Open("sqlite3", db.path)
	//debugPrint("db <%s> opend at at <%s>", db.name, db.path)
	log.Debugf("<%s> opened at <%s>", db.name, db.path)
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

	if !backupHookRegistered {
		//debugPrint("backup_hook: registering driver %s", DB_BACKUP_HOOK)
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
		backupHookRegistered = true
	}

	debugPrint("<%s> initialized", db.name)
}

func (db *DB) Close() {
	//debugPrint("Closing <%s>", db.name)
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

func (src *DB) SyncToDisk(dbpath string) error {

	if !backupHookRegistered {
		errMsg := fmt.Sprintf("%s, %s", src.path, "db backup hook is not initialized")
		return errors.New(errMsg)
	}

	//debugPrint("[flush] openeing <%s>", src.path)
	srcDb, err := sql.Open(DB_BACKUP_HOOK, src.path)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//debugPrint("[flush] opening <%s>", DB_FILENAME)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	bkDb, err := sql.Open(DB_BACKUP_HOOK, dbUri)
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

func (dst *DB) SyncFromDisk(dbpath string) error {

	if !backupHookRegistered {
		errMsg := fmt.Sprintf("%s, %s", dst.path, "db backup hook is not initialized")
		return errors.New(errMsg)
	}

	debugPrint("Syncing <%s> to <%s>", dbpath, dst.name)

	dbUri := fmt.Sprintf("file:%s", dbpath)
	srcDb, err := sql.Open(DB_BACKUP_HOOK, dbUri)
	defer flushSqliteCon(srcDb)
	if err != nil {
		return err
	}
	srcDb.Ping()

	//debugPrint("[flush] opening <%s>", DB_FILENAME)
	bkDb, err := sql.Open(DB_BACKUP_HOOK, dst.path)
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

// TODO: Use context when making call from request/api
func initDB() {

	// Initialize memory db with schema
	cacheDB = DB{}.New("memcache", DB_MEMCACHE_PATH)
	cacheDB.Init()

	// Check and initialize local db as last step
	// browser bookmarks should already be in cache

	dbdir := getDefaultDBPath()
	dbpath := filepath.Join(dbdir, DB_FILENAME)

	// Verifiy that local db directory path is writeable
	err := checkWriteable(dbdir)
	logPanic(err)

	// If local db exists load it to cacheDB
	var exists bool
	if exists, err = checkFileExists(dbpath); exists {
		logPanic(err)
		debugPrint("localdb exists, preloading to cache")
		cacheDB.SyncFromDisk(dbpath)
		//_ = cacheDB.Print()
	} else {
		logPanic(err)
		// Else initialize it
		initLocalDB(cacheDB, dbpath)
	}

}

//Initialize the database here
//Bla bla bla
func initLocalDB(db *DB, dbpath string) {

	log.Infof("Initializing local db at '%s'", dbpath)
	debugPrint("%s flushing to disk", db.name)
	err := db.SyncToDisk(dbpath)
	logPanic(err)

}

func flushSqliteCon(con *sql.DB) {
	con.Close()
	_sql3conns = _sql3conns[:len(_sql3conns)-1]
	log.Debugf("Flushed sqlite conns %v", _sql3conns)
}
