package firefox

import (
	"strings"

	"git.sp4ke.xyz/sp4ke/gomark/cmd"
	"git.sp4ke.xyz/sp4ke/gomark/config"
	"git.sp4ke.xyz/sp4ke/gomark/utils"

	"github.com/gobuffalo/flect"
	"github.com/urfave/cli/v2"
)

const (
	FirefoxProfileFlag = "firefox-profile"
)

var globalFirefoxFlags = []cli.Flag{
    // This allows us to register dynamic cli flags which get converted to 
    // config.Configurator options. 
    // The flag must be given a name in the form `--firefox-<flag>`.
	&cli.StringFlag{
		Name:  FirefoxProfileFlag,
		Usage: "Set the default firefox `PROFILE` to use",
	},
    // &cli.StringFlag{
    //     Name: "firefox-default-dir",
    //     Usage: "test",
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

		if sp[0] != "firefox" {
			continue
		}

		//TODO: document this feature
        // extract global options that start with --firefox-*
		optionName := flect.Pascalize(strings.Join(sp[1:], " "))
		var destVal interface{}

		// Find the corresponding flag
		for _, ff := range globalFirefoxFlags {
			if ff.String() == f.String() {

				// Type switch on the flag type
				switch ff.(type) {

				case *cli.StringFlag:
					destVal = c.String(f.Names()[0])
				}
			}
		}

		err := config.RegisterModuleOpt(BrowserName,
			optionName, destVal)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func init() {
	cmd.RegBeforeHook(BrowserName, globalCommandFlagsManager)

    for _, flag := range globalFirefoxFlags {
        cmd.RegGlobalFlag(BrowserName, flag)
    }
}
