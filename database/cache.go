package database

const (
	CacheName = "memcache"
	//MemcacheFmt = "file:%s?mode=memory&cache=shared"
	//BufferFmt   = "file:%s?mode=memory&cache=shared"
	DBTypeInMemoryDSN = "file:%s?mode=memory&cache=shared"
	DBTypeCacheDSN    = DBTypeInMemoryDSN
)

var (
	// Global cache database
	// Main in memory db, is synced with disc
	// `CacheDB` is a memory replica of disk db
	Cache = &CacheDB{}
)

type CacheDB struct {
	DB *DB
}

func (c *CacheDB) IsInitialized() bool {
	return Cache.DB != nil && Cache.DB.Handle != nil
}

func initCache() {
	log.Debug("initializing cacheDB")
	var err error
	// Initialize memory db with schema
	Cache.DB, err = NewDB(CacheName, "", DBTypeCacheDSN).Init()
	if err != nil {
		log.Fatal(err)
	}

	err = Cache.DB.InitSchema()
	if err != nil {
		log.Fatal(err)
	}
}
