package main

import (
	"os"

	"git.sp4ke.xyz/sp4ke/gomark/modules"
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
	// Category: "daemon"
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

	registeredBrowsers := modules.GetBrowserModules()

	// instanciate all browsers
	for _, browserMod := range registeredBrowsers {

		mod := browserMod.ModInfo()

		//Create a browser instance
		browser, ok := mod.New().(modules.BrowserModule)
		if !ok {
			log.Criticalf("module <%s> is not a BrowserModule", mod.ID)
		}
		log.Debugf("created browser instance <%s>", browser.Config().Name)

		// shutdown logic
		s, isShutdowner := browser.(modules.Shutdowner)
		if isShutdowner {
			defer handleShutdown(browser.Config().Name, s)
		}

		log.Debugf("new browser <%s> instance", browser.Config().Name)
		h, ok := browser.(modules.HookRunner)
		if ok {
			//TODO: document hook running
			h.RegisterHooks(parsing.ParseTags)
		}

		//TODO: call the setup logic for each browser instance
		// includes the browsers.Initializer and browsers.Loader interfaces
		err := modules.Setup(browser)
		if err != nil {
			log.Errorf("setting up <%s> %v", browser.Config().Name, err)
			if isShutdowner {
				handleShutdown(browser.Config().Name, s)
			}
			continue
		}

		runner, ok := browser.(watch.WatchRunner)
		if !ok {
			log.Criticalf("<%s> must implement watch.WatchRunner interface", browser.Config().Name)
			continue
		}

		log.Infof("start watching <%s>", runner.Watcher().ID)
		watch.SpawnWatcher(runner)
	}

	<-manager.Quit

	return nil
}

func handleShutdown(id string, s modules.Shutdowner) {
	err := s.Shutdown()
	if err != nil {
		log.Panicf("could not shutdown browser <%s>", id)
	}
}
