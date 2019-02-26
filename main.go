// ### API Usage:
// - Init Mode (debug, release) and Logging
// - Create a Browser object for each browser using `New[BrowserType]()`
// - Call `Load()` and `Watch()` on every browser
// - Run the gin server
package main

import (
	"gomark/cmd"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gomark"
	app.Version = "1.0"

	app.Commands = []cli.Command{
		startServerCmd,
		cmd.FirefoxCmds,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
