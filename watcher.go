package main

import (
	"io/ioutil"
	"log"

	"github.com/buger/jsonparser"
	"github.com/fsnotify/fsnotify"
)

func googleParseBookmarks(bookmarkPath string) {
	f, err := ioutil.ReadFile(BOOKMARK_FILE)

	if err != nil {
		log.Fatal(err)
	}

	debugPrint("Parsing bookmarks")
	// Begin parsing
	rootsData, _, _, _ := jsonparser.Get(f, "roots")

	jsonparser.ObjectEach(rootsData, gJsonParseRecursive)
	// Finished parsing
	debugPrint("parsed %d bookmarks", parserStat.lastUrlCount)
}

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
