package main

const (
	BOOKMARK_FILE = "Bookmarks"
	BOOKMARK_DIR  = "/home/spike/.config/google-chrome/Default/"
)

var Channels = struct {
	bookmarkWatcher chan bool
}{}

//func startWatcher() {

//watcher, err := fsnotify.NewWatcher()
//defer watcher.Close()

//go watcherThread(watcher)

//// Watch chrome bookmark dir
//err = watcher.Add(BOOKMARK_DIR)
//logPanic(err)

//<-Channels.bookmarkWatcher
//}

func main() {

	// Block the main function
	block := make(chan bool)

	// Initialize sqlite database available in global `CACHE_DB` variable
	err := initDB()
	logPanic(err)
	//debugPrint("%v", isEmptyDb(currentJobDB))
	//debugPrint("%v", isEmptyDb(memCacheDb))

	// Preload existing bookmarks
	//debugPrint("Preload bookmarks")
	//googleParseBookmarks(BOOKMARK_FILE)

	//debugPrint("%v", isEmptyDb(currentJobDB))
	//debugPrint("%v", isEmptyDb(memCacheDb))

	//printDbCount(currentJobDB)
	//printDbCount(memCacheDb)

	//debugPrint("%v", isEmptyDb(memCacheDb))

	chromeWatcher := &bookmarkWatcher{}
	chromeWatcher.Init(BOOKMARK_DIR, BOOKMARK_FILE, Chrome)
	chromeWatcher.Start()

	// Flush to disk for testing
	//flushToDisk()

	<-block
}
