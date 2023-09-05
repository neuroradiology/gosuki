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
			modinfo := mod.ModInfo()
			hook := cmd.BeforeHook(string(modinfo.ID))
			if hook != nil {
				if err := cmd.BeforeHook(string(modinfo.ID))(c); err != nil {
					return err
				}
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
		cmd.ModuleCmds,
	}

	// Add global flags from registered modules
	// we use GetModules to handle all types of modules
	modules := modules.GetModules()
	log.Debugf("loading %d modules", len(modules))
	for _, mod := range modules {
		modID := string(mod.ModInfo().ID)
		log.Debugf("loading module <%s>", modID)

		// for each registered module, register own flag management
		modFlags := cmd.GlobalFlags(modID)
		if len(modFlags) != 0 {
			app.Flags = append(app.Flags, modFlags...)
		}

		// Add all browser module registered commands
		cmds := cmd.RegisteredModCommands(modID)
		for _, cmd := range cmds {
			app.Commands = append(app.Commands, cmd)
		}
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func init() {

	//TODO: watch all profiles (handled at browser level for now)
	// config.RegisterGlobalOption("all-profiles", false)
}
