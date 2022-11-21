package cmd

import (
	"git.sp4ke.xyz/sp4ke/gomark/config"
	"git.sp4ke.xyz/sp4ke/gomark/logging"

	"github.com/kr/pretty"
	"github.com/urfave/cli/v2"
)

var log = logging.GetLogger("CMD")

var cfgPrintCmd = &cli.Command{
	Name:    "print",
	Aliases: []string{"p"},
	Usage:   "print current config",
	Action:  printConfig,
}

var ConfigCmds = &cli.Command{
	Name:  "config",
	Usage: "get/set config opetions",
	Subcommands: []*cli.Command{
		cfgPrintCmd,
	},
}

func printConfig(c *cli.Context) error {
	pretty.Println(config.GetAll())

    return nil
}
