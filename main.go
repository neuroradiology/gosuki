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
	defer db.Close()

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

	// Preload existing bookmarks
	googleParseBookmarks(BOOKMARK_FILE)

	// Flush to disk for testing
	//flushToDisk()

	<-done
}
