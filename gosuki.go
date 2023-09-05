// # Gosuki documentation
package main

import (
	"os"

	"git.blob42.xyz/gomark/gosuki/config"
	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/modules"

	"git.blob42.xyz/gomark/gosuki/cmd"

	"github.com/urfave/cli/v2"

	// Load firefox browser modules
	_ "git.blob42.xyz/gomark/gosuki/firefox"

	// Load chrome browser module
	_ "git.blob42.xyz/gomark/gosuki/chrome"
)

func main() {

	app := cli.NewApp()


	app.Name = "gosuki"
	app.Version = version()

	flags := []cli.Flag{

		&cli.StringFlag{
			Name:  "config-file",
			Value: config.ConfigFile,
			Usage: "TOML config `FILE` path",
		},

        &cli.IntFlag{
        	Name:        "debug",
        	Aliases:     []string{"d"},
        	EnvVars:     []string{logging.EnvGosukiDebug},
            Action: func (c *cli.Context, val int) error {
                logging.SetMode(val)
                return nil
            },

        },
	}

	flags = append(flags, config.SetupGlobalFlags()...)
	app.Flags = append(app.Flags, flags...)

	app.Before = func(c *cli.Context) error {


		// get all registered browser modules
		modules := modules.GetModules()
		for _, mod := range modules {

			// Run module's hooks that should run before context is ready
			// for example setup flags management
			err := cmd.BeforeHook(string(mod.ModInfo().ID))(c)
			if err != nil {
				return err
			}
		}

		// Execute config hooks
		//TODO: better doc for what are Conf hooks ???
		config.RunConfHooks(c)

		initConfig()

		return nil
	}


	// Browser modules can register commands through cmd.RegisterModCommand.
	// registered commands will be appended here
	app.Commands = []*cli.Command{
		// main entry point
		startDaemonCmd,
		cmd.ConfigCmds,
		cmd.ProfileCmds,
	}

	// Add global flags from registered modules
	// we use GetModules to handle all types of modules
	modules := modules.GetModules()
	log.Debugf("loading %d modules", len(modules))
	for _, mod := range modules {
		modId := string(mod.ModInfo().ID)
		log.Debugf("loading module <%s>", modId)

		// for each registered module, register own flag management
		mod_flags := cmd.GlobalFlags(modId)
		if len(mod_flags) != 0 {
			app.Flags = append(app.Flags, mod_flags...)
		}

		// Add all browser module registered commands
		cmds := cmd.ModCommands(modId)
		for _, cmd := range cmds {
			app.Commands = append(app.Commands, cmd)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func init() {
	config.RegisterGlobalOption("watch-all-profiles", false)
}
