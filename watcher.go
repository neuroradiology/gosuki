package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func watcherThread(watcher *fsnotify.Watcher) {
	for {
		select {
		case event := <-watcher.Events:

			if event.Op&fsnotify.Create == fsnotify.Create &&
				event.Name == BOOKMARK_FILE {

				debugPrint("event: %v| name: %v", event.Op, event.Name)
				debugPrint("modified file:", event.Name)
				googleParseBookmarks(BOOKMARK_FILE)

			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
