package main

import (
	"gomark/mozilla"
	"gomark/parsing"
	"gomark/utils"
	"os"
	"path"

	"git.sp4ke.com/sp4ke/gum"
	"github.com/urfave/cli"
)

var startServerCmd = cli.Command{
	Name:    "server",
	Aliases: []string{"s"},
	Usage:   "run browser watchers",
	Action:  startServer,
}

var unlockFirefoxCmd = cli.Command{
	Name:   "ff",
	Usage:  "Disable VFS lock in firefox",
	Action: unlockFirefox,
}

func unlockFirefox(c *cli.Context) {
	prefsPath := path.Join(mozilla.BookmarkDir, mozilla.PrefsFile)

	pusers, err := utils.FileProcessUsers(path.Join(mozilla.BookmarkDir, mozilla.BookmarkFile))
	if err != nil {
		fflog.Error(err)
	}

	for pid, p := range pusers {
		pname, err := p.Name()
		if err != nil {
			fflog.Error(err)
		}
		log.Errorf("multiprocess not enabled and %s(%d) is running. Close firefox and disable VFS lock", pname, pid)
	}
	// End testing

	// enable multi process access in firefox
	err = mozilla.SetPrefBool(prefsPath,
		mozilla.PrefMultiProcessAccess,
		true)

	if err != nil {
		log.Error(err)
	}
}

func startServer(c *cli.Context) {
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

	cr := NewChromeBrowser()
	if cr != nil {
		browsers = append(browsers, cr)
	}

	for _, b := range browsers {
		defer b.Shutdown()
		b.RegisterHooks(parsing.ParseTags)
		b.Load()
		b.Watch()
	}

	<-manager.Quit
}
