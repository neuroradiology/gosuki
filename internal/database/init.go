package database

import (
	"path/filepath"

	"git.blob42.xyz/gomark/gosuki/internal/utils"
)

func InitDB() {
	var err error

	// Check and initialize local db as last step
	// browser bookmarks should already be in cache

	dbdir := utils.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
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
		er := Cache.DB.SyncFromDisk(dbpath)
		if er != nil {
			log.Critical(er)
		}
	} else {
		if err != nil {
			log.Error(err)
		}

		// Else initialize it
		initLocalDB(Cache.DB, dbpath)
	}

}

// Initialize the local database file
func initLocalDB(db *DB, dbpath string) {

	log.Infof("Initializing local db at '%s'", dbpath)
	err := db.SyncToDisk(dbpath)
	if err != nil {
		log.Critical(err)
	}

}
