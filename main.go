// ### API Usage:
// - Init Mode (debug, release) and Logging
// - Create a Browser object for each browser using `New[BrowserType]()`
// - Call `Load()` and `Watch()` on every browser
// - Run the gin server
package main

import (
	"os"

	"git.sp4ke.com/sp4ke/gomark/config"

	"git.sp4ke.com/sp4ke/gomark/cmd"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "gomark"
	app.Version = "1.0"

	flags := []cli.Flag{

		&cli.StringFlag{
			Name:  "config-file",
			Value: config.ConfigFile,
			Usage: "TOML config `FILE` path",
		},
	}

	app.Before = func(c *cli.Context) error {

		err := cmd.GlobalFirefoxFlagsManager(c)
		if err != nil {
			return err
		}

		// Execute config hooks
		config.RunConfHooks()

		return nil
	}

	app.Flags = flags
	for _, f := range cmd.FirefoxGlobalFlags {
		app.Flags = append(app.Flags, f)
	}

	app.Commands = []*cli.Command{
		startServerCmd,
		cmd.FirefoxCmds,
		cmd.ConfigCmds,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
