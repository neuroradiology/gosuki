// Modules can register custom hooks here that will plug into urfave *cli.App
// API. The hooks will be called in the same order as defined urfave's cli.
package cmd

import "github.com/urfave/cli/v2"

type Hook func(*cli.Context) error

// Map module id to list of *cli.App.Before hooks
var modCmdBeforeHooks = map[string]Hook{}

// Register a module hook to be run in *cli.App.Before
func RegBeforeHook(modId string, hook Hook) {
	if hook == nil {
		log.Panicf("cannot register nil hook for <%s>", modId)
	}

	if _, ok := modCmdBeforeHooks[modId]; ok {
		log.Warningf("a hook was already registered for module <%s>", modId)
	}
	modCmdBeforeHooks[modId] = hook
}

// Return all registered Before hooks for module
func BeforeHook(modId string) Hook {
	return modCmdBeforeHooks[modId]
}
