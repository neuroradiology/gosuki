package main

import (
	"os"

	"git.blob42.xyz/gomark/gosuki/modules"
	"git.blob42.xyz/gomark/gosuki/parsing"
	"git.blob42.xyz/gomark/gosuki/profiles"
	"git.blob42.xyz/gomark/gosuki/utils"
	"git.blob42.xyz/gomark/gosuki/watch"

	"git.blob42.xyz/sp4ke/gum"

	"github.com/urfave/cli/v2"
)

var startDaemonCmd = &cli.Command{
	Name:    "daemon",
	Aliases: []string{"d"},
	Usage:   "run browser watchers",
	// Category: "daemon"
	Action:  startDaemon,
}

func startDaemon(c *cli.Context) error {
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

		// Create context
		modContext := &modules.Context{
			Cli: c,
		}

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
			//TODO: document hook running on watch events
			h.RegisterHooks(parsing.ParseTags)
		}

		//WIP: Handle multiple profiles for modules who announce it - here ?
		// Check if browser implements ProfileManager
		//TODO: global flag for watch all
		// if watch-all then for each profile setup the browser
		bpm, ok := browser.(profiles.ProfileManager)
		if ok {
			//TODO : for each profile spawn a watcher
			// list profiles
			profs, err := bpm.GetProfiles()
			if err != nil {
				log.Critical("could not get profiles")
			}
			for _, p := range profs {
				log.Debugf("profile: <%s>", p.Name)
			}
		} else {
			log.Debugf("<%s> does not implement profiles.ProfileManager",
			browser.Config().Name)
		}


		// calls the setup logic for each browser instance which
		// includes the browsers.Initializer and browsers.Loader interfaces
		err := modules.Setup(browser, modContext)
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

		log.Infof("start watching <%s>", runner.Watch().ID)
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
