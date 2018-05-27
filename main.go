// ###Gomark API Documentation
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

	cb.RegisterHooks(ParseTags)

	cb.Load()

	_ = cb.Watch()

	r.Run("127.0.0.1:4242")

	//<-block
}
