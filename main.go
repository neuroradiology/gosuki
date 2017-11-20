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

	cb := NewChromeBrowser()
	cb.Load()
	_ = cb.Watch()

	// Flush to disk for testing
	//flushToDisk()

	//var chrome Browser
	//chrome = browsers.New(browsres.TChrome)
	//chrome.Watch()

	<-block
}
