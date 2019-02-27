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
	"strings"

	altsrc "github.com/urfave/cli/altsrc"
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
		//log.Warning(c.GlobalString("firefox.DefaultProfile"))

		//TODO: check altsrc how to parse subsection for options
		for _, conf := range c.GlobalFlagNames() {

			// Check if this is a submodule option
			sp := strings.Split(conf, ".")
			// Submodule options
			if len(sp) > 1 {
				log.Warning(conf)
				log.Critical(config.GetAll())

				module := sp[0]
				option := sp[1]

				// find the flag
				for _, f := range flags {
					if f.GetName() == conf {

						// Type switch

						switch f.(type) {

						// String flags
						case *altsrc.StringFlag:
							//log.Criticalf("%s is String flag", conf)
							//log.Criticalf("%s, %s, %s", module, option, c.GlobalString(conf))

							config.RegisterModuleOpt(module,
								option, c.GlobalString(conf))

						}
					}
				}

			}

		}

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
