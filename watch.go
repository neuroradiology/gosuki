package main

// func watchLoop() {
//
// 	// Initialize sqlite database available in global `cacheDB` variable
// 	initDB()
//
//     // Create instance of browsers
//
//     //TODO: replace NewXX by an abstract method that allows any number of
//     //custom browsers to be registered
//
//     //TODO: get list of registed browsers by id
// 	browsers := []IBrowser{
// 		NewFFBrowser(),
// 		NewChromeBrowser(),
// 	}
//
// 	for _, b := range browsers {
// 		defer b.Shutdown()
// 		b.RegisterHooks(parsing.ParseTags)
// 		b.Load()
// 		b.Watch()
// 	}
//
// 	//cb := NewChromeBrowser()
// 	//ff := NewFFBrowser()
// 	//defer cb.Shutdown()
// 	//defer ff.Shutdown()
//
// 	//cb.RegisterHooks(parsing.ParseTags)
// 	//cb.Load()
// 	//ff.Load()
//
// 	//_ = cb.Watch()
// 	//_ = ff.Watch()
//
// }
