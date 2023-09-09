package cmd

import (
	"git.blob42.xyz/gomark/gosuki/internal/config"
	"git.blob42.xyz/gomark/gosuki/internal/logging"

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

func printConfig(_ *cli.Context) error {
	pretty.Println(config.GetAll())

    return nil
}
