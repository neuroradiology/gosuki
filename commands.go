package main

import (
	"os"

	"git.sp4ke.com/sp4ke/gomark/parsing"
	"git.sp4ke.com/sp4ke/gomark/utils"

	"git.sp4ke.com/sp4ke/gum"

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

	var browsers []IBrowser

	ff := NewFFBrowser()
	if ff != nil {
		browsers = append(browsers, ff)
	}

	//cr := NewChromeBrowser()
	//if cr != nil {
	//browsers = append(browsers, cr)
	//}

	for _, b := range browsers {
		defer b.Shutdown()
		b.RegisterHooks(parsing.ParseTags)

		err := b.Init()
		if err != nil {
			log.Criticalf("<%s> %s", b, err)
			b.Shutdown()
			continue
		}

		err = b.Load()
		if err != nil {
			log.Criticalf("<%s> %s", b, err)
			b.Shutdown()
			continue
		}

		b.Watch()
	}

	<-manager.Quit

	return nil
}
