// commands related to modules
package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"git.blob42.xyz/gomark/gosuki/modules"
)

// map cmd Name to *cli.Command
type modCmds map[string]*cli.Command

var (
	// Map registered browser module IDs to their modCmds map
	modCommands = map[string]modCmds{}
)

// TODO: use same logic with browser mod registering
func RegisterModCommand(modID string, cmd *cli.Command) {
	if cmd == nil {
		log.Panicf("cannot register nil cmd for <%s>", modID)
	}

	if _, ok := modCommands[modID]; !ok {
		modCommands[modID] = make(modCmds)
	}
	modCommands[modID][cmd.Name] = cmd
}

// return list of registered commands for browser module
func RegisteredModCommands(modID string) modCmds {
	return modCommands[modID]
}

var ModuleCmds = &cli.Command {
	Name: "modules",
	Aliases: []string{"m"},
	Usage: "module commands",
	Subcommands: []*cli.Command{
		listModulesCmd,
	},
}

var listModulesCmd = &cli.Command{
	Name: "list",
	Usage: "list available browsers and modules",
	Action: func(_ *cli.Context) error {

		fmt.Printf("\n%s\n", "Modules:")
		mods := modules.GetModules()
		for _, mod := range mods {
			_, isBrowser := mod.(modules.BrowserModule)
			if isBrowser {
				fmt.Printf("-%-10s \t %s\n", mod.ModInfo().ID, "<browser>")
			} else {
				fmt.Printf("-%-10s \t %s\n", mod.ModInfo().ID, "<module>")
			}
		}
		return nil
	},
}

