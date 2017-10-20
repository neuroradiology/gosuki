package main

import (
	"log"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"encoding/json"
)


func watcherThread(watcher *fsnotify.Watcher){
	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Create == fsnotify.Create &&
			   event.Name == BOOKMARK_FILE {
				log.Printf("event: %v| name: %v", event.Op, event.Name)
				log.Println("modified file:", event.Name)

				log.Println("Parsing bookmark")

				f, err := ioutil.ReadFile(BOOKMARK_FILE)

				if err != nil {
					log.Fatal(err)
				}

				rootData := RootData{}

				_ = json.Unmarshal(f, &rootData)

				for _, root := range rootData.Roots {
					parseJsonNodes(&root)
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}

