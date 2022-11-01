package main

import (
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

	//TODO: instanciate all browsers

	for _, browserMod := range registeredBrowsers {

		mod := browserMod.ModInfo()

		if mod.New == nil {
			log.Criticalf("browser module <%s> has no constructor", mod.ID)
			continue
		}

		// Recover the browser instance
		browser, ok := mod.New().(browsers.BrowserModule)
		if !ok {
			log.Criticalf("module <%s> is not a BrowserModule", mod.ID)
		}
		log.Debugf("created browser instance <%s>", browser.Config().Name)

		//TIP: shutdown logic
		s, isShutdowner := browser.(browsers.Shutdowner)
		if isShutdowner {
			defer handleShutdown(browser.Config().Name, s)
		}

		log.Debugf("new browser <%s> instance", browser.Config().Name)
		h, ok := browser.(browsers.HookRunner)
		if ok {
			//TODO: document hook running
			h.RegisterHooks(parsing.ParseTags)
		}

		//TODO: call the setup logic for init,load for each browser instance
		err := browsers.Setup(browser)
		if err != nil {
			log.Errorf("setting up <%s> %v", browser.Config().Name, err)
			if isShutdowner {
				handleShutdown(browser.Config().Name, s)
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

		runner, ok := browser.(watch.WatchRunner)
		if !ok {
			log.Criticalf("<%s> must implement watch.WatchRunner interface", browser.Config().Name)
			continue
		}

		watch.SpawnWatcher(runner)
		// b.Browser.Watch()
	}

	<-manager.Quit

	return nil
}

func handleShutdown(id string, s browsers.Shutdowner) {
	err := s.Shutdown()
	if err != nil {
		log.Panicf("could not shutdown browser <%s>", id)
	}
}
