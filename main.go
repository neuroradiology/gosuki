package main

const (
	BOOKMARK_FILE = "Bookmarks"
	BOOKMARK_DIR  = "/home/spike/.config/google-chrome/Default/"
)

func main() {

	// Block the main function
	block := make(chan bool)

	initMode()
	initLogging()
	// Initialize sqlite database available in global `cacheDB` variable
	initDB()

	chromeWatcher := &bookmarkWatcher{}
	chromeWatcher.Init(BOOKMARK_DIR, BOOKMARK_FILE, Chrome)
	chromeWatcher.Preload()
	chromeWatcher.Start()

	// Flush to disk for testing
	//flushToDisk()

	<-block
}
