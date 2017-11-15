package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

const (
	BOOKMARK_FILE = "/home/spike/.config/google-chrome/Default/Bookmarks"
	BOOKMARK_DIR  = "/home/spike/.config/google-chrome/Default/"
)

func main() {

	// Initialize sqlite database available in global `db` variable
	initDB()
	defer memCacheDb.Close()
	defer currentJobDB.Close()
	//debugPrint("%v", isEmptyDb(currentJobDB))
	//debugPrint("%v", isEmptyDb(memCacheDb))

	// Preload existing bookmarks
	debugPrint("Preload bookmarks")
	googleParseBookmarks(BOOKMARK_FILE)

	//debugPrint("%v", isEmptyDb(currentJobDB))
	//debugPrint("%v", isEmptyDb(memCacheDb))

	//printDbCount(currentJobDB)
	//printDbCount(memCacheDb)

	//debugPrint("%v", isEmptyDb(memCacheDb))

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go watcherThread(watcher)

	// Watch chrome bookmark dir
	err = watcher.Add(BOOKMARK_DIR)
	if err != nil {
		log.Fatal(err)
	}

	// Flush to disk for testing
	//flushToDisk()

	<-done
}
