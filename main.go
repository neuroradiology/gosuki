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

	altsrc "github.com/urfave/cli/altsrc"
	cli "gopkg.in/urfave/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "gomark"
	app.Version = "1.0"

	flags := []cli.Flag{

		cli.StringFlag{
			Name:  "config",
			Value: config.ConfigFile,
			Usage: "TOML config `FILE` path",
		},
	}

	app.Before = func(c *cli.Context) error {

		err := altsrc.InitInputSourceWithContext(flags,
			altsrc.NewTomlSourceFromFlagFunc("config"))(c)
		if err != nil {
			return err
		}

		//TODO: check altsrc how to parse subsection for options
		//for _, conf := range c.GlobalFlagNames() {

		//log.Debug(conf)
		//err := config.RegisterConf(flect.Pascalize(conf), c.GlobalString(conf))
		//if err != nil {
		//return err
		//}

		//}

		return nil
	}

	app.Flags = flags

	app.Commands = []cli.Command{
		startServerCmd,
		cmd.FirefoxCmds,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
