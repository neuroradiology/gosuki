package main

import (
	"fmt"
	"os"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"git.sp4ke.xyz/sp4ke/gomark/watch"

	"git.sp4ke.xyz/sp4ke/gum"

	"github.com/urfave/cli/v2"
)

var startServerCmd = &cli.Command{
	Name:    "server",
	Aliases: []string{"s"},
	Usage:   "run browser watchers",
	Action:  startServer,
}

func startServer(c *cli.Context) error {
	defer utils.CleanFiles()
	manager := gum.NewManager()
	manager.ShutdownOn(os.Interrupt)

	api := NewApi()
	manager.AddUnit(api)

	go manager.Run()

	// Initialize sqlite database available in global `cacheDB` variable
	initDB()

	registeredBrowsers := browsers.Modules()
	log.Debugf("registered browsers: %v", registeredBrowsers)

	//TODO: instanciate all browsers

	for _, mod := range registeredBrowsers {

		// type assert to Browser interface
		var browser browsers.BrowserModule
		browser, ok := mod.(browsers.BrowserModule)
		if !ok {
			log.Errorf("<%s> is not a browser module", mod.ModInfo().ID)
			continue
		}

		//TIP: shutdown logic
		// shutdowner, isShutdowner := browser.(browsers.Shutdowner)
		// if isShutdowner {
		// 	defer shutdowner.Shutdown()
		// }

		log.Debugf("new browser instance with path %s", browser.Config().BookmarkPath())
		h, ok := browser.(browsers.HookRunner)
		if ok {
			//TODO: document hook running
			h.RegisterHooks(parsing.ParseTags)
		}

		//TODO: call the setup logic for init,load for each browser instance
		err := browsers.Setup(browser)
		if err != nil {
			log.Criticalf("setting up <%s> %v", browser.Config().Name, err)
			if isShutdowner {
				err := shutdowner.Shutdown()
				if err != nil {
					log.Critical(fmt.Errorf("shutting down <%s>: %v", browser.Config().Name, err))
				}
			}
			continue
		}

		// err := b.Init()
		// if err != nil {
		// 	log.Criticalf("<%s> %s", b, err)
		// 	b.Shutdown()
		// 	continue
		// }
		//
		// err = b.Load()
		// if err != nil {
		// 	log.Criticalf("<%s> %s", b, err)
		// 	b.Shutdown()
		// 	continue
		// }

		watch.SpawnWatcher(browser)
		// b.Browser.Watch()
	}

	<-manager.Quit

	return nil
}
