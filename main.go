// ### API Usage:
// - Init Mode (debug, release) and Logging
// - Create a Browser object for each browser using `New[BrowserType]()`
// - Call `Load()` and `Watch()` on every browser
// - Run the gin server
package main

import "github.com/gin-gonic/gin"

func main() {
	// Block the main function
	//block := make(chan bool)

	r := gin.Default()

	r.GET("/urls", getBookmarks)

	initMode()
	initLogging()

	// Initialize sqlite database available in global `cacheDB` variable
	initDB()

	cb := NewChromeBrowser()
	ff := NewFFBrowser()

	cb.RegisterHooks(ParseTags)
	ff.RegisterHooks(ParseTags)

	cb.Load()
	ff.Load()

	_ = cb.Watch()
	_ = ff.Watch()

	r.Run("127.0.0.1:4242")

	//<-block
}
