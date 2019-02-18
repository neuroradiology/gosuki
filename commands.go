package main

import (
	"gomark/parsing"
	"os"

	"git.sp4ke.com/sp4ke/gum"
	"github.com/urfave/cli"
)

var startServerCmd = cli.Command{
	Name:    "server",
	Aliases: []string{"s"},
	Usage:   "run browser watchers",
	Action:  startServer,
}

func startServer(c *cli.Context) {
	manager := gum.NewManager()
	manager.ShutdownOn(os.Interrupt)

	api := NewApi()
	manager.AddUnit(api)

	go manager.Run()

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

	<-manager.Quit
}
