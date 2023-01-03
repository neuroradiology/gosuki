package cmd

import (
	"github.com/urfave/cli/v2"
)

// map cmd Name to *cli.Command
type modCmds map[string]*cli.Command

var (
	// Map browser module IDs to their modCmds map
	modCommands = map[string]modCmds{}
)

// TODO: use same logic with browser mod registering
func RegisterModCommand(modId string, cmd *cli.Command) {
	if cmd == nil {
		log.Panicf("cannot register nil cmd for <%s>", modId)
	}

	if _, ok := modCommands[modId]; !ok {
		modCommands[modId] = make(modCmds)
	}
	modCommands[modId][cmd.Name] = cmd
}

// return list of registered commands for browser module
func ModCommands(modId string) modCmds {
	return modCommands[modId]
}
