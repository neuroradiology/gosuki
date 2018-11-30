package main

import (
	"fmt"
	"gomark/database"
	"gomark/utils"
	"path/filepath"
)

type DB = database.DB

// Global cache database
var (
	CacheDB *DB // Main in memory db, is synced with disc
)

func initDB() {
	var err error
	// Initialize memory db with schema
	cachePath := fmt.Sprintf(database.DBMemcacheFmt, database.DBCacheName)
	CacheDB, err = database.New(database.DBCacheName, cachePath)
	log.Debugf("cache %#v", CacheDB)
	if err != nil {
		log.Critical(err)
	}

	// Check and initialize local db as last step
	// browser bookmarks should already be in cache

	dbdir := utils.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, database.DBFileName)
	// Verifiy that local db directory path is writeable
	err = utils.CheckWriteable(dbdir)
	if err != nil {
		log.Critical(err)
	}

	// If local db exists load it to cacheDB
	var exists bool
	if exists, err = utils.CheckFileExists(dbpath); exists {
		if err != nil {
			log.Warning(err)
		}
		log.Infof("<%s> exists, preloading to cache", dbpath)
		er := CacheDB.SyncFromDisk(dbpath)
		if er != nil {
			log.Critical(er)
		}
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
