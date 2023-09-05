package firefox

import (
	"strings"

	"git.blob42.xyz/gomark/gosuki/cmd"
	"git.blob42.xyz/gomark/gosuki/config"
	"git.blob42.xyz/gomark/gosuki/utils"

	"github.com/gobuffalo/flect"
	"github.com/urfave/cli/v2"
)

const (
	FirefoxProfileFlag = "ff-profile"
)

var globalFirefoxFlags = []cli.Flag{
    // This allows us to register dynamic cli flags which get converted to 
    // config.Configurator options. 
    // The flag must be given a name in the form `--firefox-<flag>`.
	&cli.StringFlag{
		Name:  FirefoxProfileFlag,
		Category: "firefox",
		Usage: "Set the default firefox `PROFILE` to use",
	},
	// &cli.BoolFlag{
	// 	Name:        "ff-watch-all-profiles",
	// 	Category:	 "firefox",
	// 	Usage:       "Watch all detected firefox profiles at the same time.",
	// 	Aliases:     []string{"ff-watch-all"},
	// },
}

// Firefox global flags must start with --firefox-<flag name here>
// NOTE: is called in *cli.App.Before callback
func globalCommandFlagsManager(c *cli.Context) error {
	log.Debugf("<%s> registering global flag manager", BrowserName)
	for _, f := range c.App.Flags {

		if utils.InList(f.Names(), "help") ||
			utils.InList(f.Names(), "version") {
			continue
		}

		if !c.IsSet(f.Names()[0]) {
			continue
		}

		sp := strings.Split(f.Names()[0], "-")

		if len(sp) < 2 {
			continue
		}

		// TEST:
		// TODO: add doc
		// Firefox flags must start with --firefox-<flag name here>
		// or -ff-<flag name here>
		if !utils.InList([]string{"firefox", "ff"}, sp[0]) {
			continue
		}

		//TODO: document this feature
        // extracts global options that start with --firefox-*
		optionName := flect.Pascalize(strings.Join(sp[1:], " "))
		var destVal interface{}

		// Find the corresponding flag
		for _, ff := range globalFirefoxFlags {
			if ff.String() == f.String() {

				// Type switch on the flag type
				switch ff.(type) {

				case *cli.StringFlag:
					destVal = c.String(f.Names()[0])

				case *cli.BoolFlag:
					destVal = c.Bool(f.Names()[0])
				}

			}
		}

		err := config.RegisterModuleOpt(BrowserName,
			optionName, destVal)

		if err != nil {
			log.Panic(err)
		}
	}
	return nil
}

func init() {
	// register dynamic flag manager for firefox
	cmd.RegBeforeHook(BrowserName, globalCommandFlagsManager)

    for _, flag := range globalFirefoxFlags {
        cmd.RegGlobalModFlag(BrowserName, flag)
    }
}
