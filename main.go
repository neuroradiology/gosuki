// ### API Usage:
// - Init Mode (debug, release) and Logging
// - Create a Browser object for each browser using `New[BrowserType]()`
// - Call `Load()` and `Watch()` on every browser
// - Run the gin server
package main

import (
	"gomark/parsing"

	"github.com/gin-gonic/gin"
)

func mainLoop() {

	r := gin.Default()

	r.GET("/urls", getBookmarks)

	// Initialize sqlite database available in global `cacheDB` variable
	initDB()

	browsers := []IBrowser{
		NewFFBrowser(),
		NewChromeBrowser(),
	}

	for _, b := range browsers {
		defer b.Shutdown()
		b.RegisterHooks(parsing.ParseTags)
		b.Load()
		b.Watch()
	}

	//cb := NewChromeBrowser()
	//ff := NewFFBrowser()
	//defer cb.Shutdown()
	//defer ff.Shutdown()

	//cb.RegisterHooks(parsing.ParseTags)
	//cb.Load()
	//ff.Load()

	//_ = cb.Watch()
	//_ = ff.Watch()

	err := r.Run("127.0.0.1:4242")
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	mainLoop()
}
