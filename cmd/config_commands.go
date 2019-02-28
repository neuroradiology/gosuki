package cmd

import (
	"gomark/config"
	"gomark/logging"
	"gomark/utils"

	cli "gopkg.in/urfave/cli.v1"
)

var log = logging.GetLogger("CMD")

var cfgPrintCmd = cli.Command{
	Name:    "print",
	Aliases: []string{"p"},
	Usage:   "print current config",
	Action:  printConfig,
}

var ConfigCmds = cli.Command{
	Name:  "config",
	Usage: "get/set config opetions",
	Subcommands: []cli.Command{
		cfgPrintCmd,
	},
}

func printConfig(c *cli.Context) {
	err := utils.PrettyPrint(config.GetAll())
	if err != nil {
		log.Fatal(err)
	}

}
