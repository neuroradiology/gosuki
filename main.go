// ### API Usage:
// - Init Mode (debug, release) and Logging
// - Create a Browser object for each browser using `New[BrowserType]()`
// - Call `Load()` and `Watch()` on every browser
// - Run the gin server
package main

import (
	"gomark/cmd"
	"gomark/config"
	"os"

	"github.com/urfave/cli/altsrc"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "gomark"
	app.Version = "1.0"

	flags := []cli.Flag{
		altsrc.NewStringFlag(cli.StringFlag{
			Name: "firefox.DefaultProfile",
		}),

		cli.StringFlag{
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
		//log.Warning(c.GlobalString("firefox.DefaultProfile"))

		// Execute config hooks
		config.RunConfHooks()

		return nil
	}

	app.Flags = flags
	for _, f := range cmd.FirefoxGlobalFlags {
		app.Flags = append(app.Flags, f)
	}

	app.Commands = []cli.Command{
		startServerCmd,
		cmd.FirefoxCmds,
		cmd.ConfigCmds,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
