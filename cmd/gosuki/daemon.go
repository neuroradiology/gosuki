package main

import (
	"fmt"
	"os"

	"git.blob42.xyz/gomark/gosuki/internal/api"
	db "git.blob42.xyz/gomark/gosuki/internal/database"
	"git.blob42.xyz/gomark/gosuki/pkg/modules"
	"git.blob42.xyz/gomark/gosuki/pkg/profiles"
	"git.blob42.xyz/gomark/gosuki/internal/utils"
	"git.blob42.xyz/gomark/gosuki/pkg/watch"

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

// Runs the module by calling the setup 
func runModule(c *cli.Context,
				browserMod modules.BrowserModule,
				p *profiles.Profile) (error) {
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
		if !isShutdowner {
			log.Warningf("browser <%s> does not implement modules.Shutdowner", browser.Config().Name)
		}

		log.Debugf("new browser <%s> instance", browser.Config().Name)


		//TODO!: call with custom profile
		if p != nil {
			bpm, ok := browser.(profiles.ProfileManager)
			if !ok {
				err := fmt.Errorf("<%s> does not implement profiles.ProfileManager",
				browser.Config().Name)
				log.Critical(err)
				return err
			}
			if err := bpm.UseProfile(*p); err != nil {
				log.Criticalf("could not use profile <%s>", p.Name)
				return err
			}
		}


		// calls the setup logic for each browser instance which
		// includes the browsers.Initializer and browsers.Loader interfaces
		err := modules.Setup(browser, modContext)
		if err != nil {
			log.Errorf("setting up <%s> %v", browser.Config().Name, err)
			if isShutdowner {
				handleShutdown(browser.Config().Name, s)
			}
			return err
		}

		runner, ok := browser.(watch.WatchRunner)
		if !ok {
			err =  fmt.Errorf("<%s> must implement watch.WatchRunner interface", browser.Config().Name)
			log.Critical(err)
			return err
		}

		log.Infof("start watching <%s>", runner.Watch().ID)
		watch.SpawnWatcher(runner)
		return nil
}

func startDaemon(c *cli.Context) error {
	defer utils.CleanFiles()
	manager := gum.NewManager()
	manager.ShutdownOn(os.Interrupt)

	api := api.NewApi()
	manager.AddUnit(api)

	go manager.Run()

	// Initialize sqlite database available in global `cacheDB` variable
	db.InitDB()

	registeredBrowsers := modules.GetBrowserModules()

	// instanciate all browsers
	for _, browserMod := range registeredBrowsers {

		mod := browserMod.ModInfo()

		//Create a temporary browser instance to check if it implements
		// the ProfileManager interface
		browser, ok := mod.New().(modules.BrowserModule)
		if !ok {
			log.Criticalf("module <%s> is not a BrowserModule", mod.ID)
		}

		// if the module is a profile manager and is watching all profiles
		// call runModule for each profile
		bpm, ok := browser.(profiles.ProfileManager)
		if ok {
			if bpm.WatchAllProfiles() {
				profs, err := bpm.GetProfiles()
				if err != nil {
					log.Critical("could not get profiles")
					continue
				}
				for _, p := range profs {
					log.Debugf("profile: <%s>", p.Name)
					err = runModule(c, browserMod, p)
					if err != nil {
					  continue
					}
				}
			} else {
				err := runModule(c, browserMod, nil)
				if err != nil {
				  continue
				}
			}
		} else {
			log.Warningf("module <%s> does not implement profiles.ProfileManager",
			browser.Config().Name)
			if err := runModule(c, browserMod, nil); err != nil {
				continue
			}
		}

		// register defer shutdown logic
		s, isShutdowner := browser.(modules.Shutdowner)
		if isShutdowner {
			defer handleShutdown(browser.Config().Name, s)
		}
	}

	<-manager.Quit

	return nil
}

func handleShutdown(id string, s modules.Shutdowner) {
	err := s.Shutdown()
	if err != nil {
		log.Panicf("<%s> could not shutdown: %s", id, err)
	}
}
