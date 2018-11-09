package main

import (
	"fmt"
	"gomark/database"
	"gomark/tools"
	"path/filepath"

	sqlite3 "github.com/mattn/go-sqlite3"
)

type DB = database.DB

// Global cache database
var (
	CacheDB              *DB                   // Main in memory db, is synced with disc
	_sql3conns           []*sqlite3.SQLiteConn // Only used for backup hook
	backupHookRegistered bool                  // set to true once the backup hook is registered
)

// TODO: Use context when making call from request/api
func initDB() {
	// Initialize memory db with schema
	cachePath := fmt.Sprintf(database.DBMemcacheFmt, database.DBCacheName)
	CacheDB = DB{}.New(database.DBCacheName, cachePath)
	CacheDB.Init()

	// Check and initialize local db as last step
	// browser bookmarks should already be in cache

	dbdir := tools.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, database.DBFileName)
	// Verifiy that local db directory path is writeable
	err := tools.CheckWriteable(dbdir)
	if err != nil {
		log.Critical(err)
	}

	// If local db exists load it to cacheDB
	var exists bool
	if exists, err = tools.CheckFileExists(dbpath); exists {
		if err != nil {
			log.Warning(err)
		}
		log.Infof("<%s> exists, preloading to cache", dbpath)
		CacheDB.SyncFromDisk(dbpath)
		//CacheDB.Print()
	} else {
		if err != nil {
			log.Error(err)
		}

		// Else initialize it
		initLocalDB(CacheDB, dbpath)
	}

}

//Initialize the local database file
func initLocalDB(db *DB, dbpath string) {

	log.Infof("Initializing local db at '%s'", dbpath)
	err := db.SyncToDisk(dbpath)
	if err != nil {
		log.Critical(err)
	}

}
