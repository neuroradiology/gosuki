package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
)

const (
	BOOKMARK_FILE="/home/spike/.config/google-chrome-unstable/Default/Bookmarks"
	BOOKMARK_DIR="/home/spike/.config/google-chrome-unstable/Default/"
)

func main() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go watcherThread(watcher)



	err = watcher.Add(BOOKMARK_DIR)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
