package main

import (
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
)

func watchLoop() {

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

}
