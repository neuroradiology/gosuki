package main

import (
	"os"

	"git.sp4ke.xyz/sp4ke/gomark/browsers"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/utils"

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

	for _, b := range registeredBrowsers {
		defer b.Browser.Shutdown()
		log.Debugf("new browser instance with path %s", b.Browser.GetBookmarksPath())
		b.Browser.RegisterHooks(parsing.ParseTags)

		//TODO: call the setup logic for init,load for each browser instance
		err := browsers.Setup(b.Browser)
		if err != nil {
			log.Criticalf("<%s> %s", b, err)
			b.Browser.Shutdown()
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

		b.Browser.Watch()
	}

	<-manager.Quit

	return nil
}
