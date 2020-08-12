package cmd

import (
	"strings"

	"git.sp4ke.com/sp4ke/gomark/config"
	"git.sp4ke.com/sp4ke/gomark/mozilla"

	"github.com/gobuffalo/flect"
	"github.com/urfave/cli/v2"
)

const (
	FirefoxDefaultProfileFlag = "firefox-default-profile"
)

var FirefoxGlobalFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  FirefoxDefaultProfileFlag,
		Usage: "Set the default firefox `PROFILE` to use",
	},
}

func GlobalFirefoxFlagsManager(c *cli.Context) error {
	flags := c.GlobalFlagNames()
	for _, f := range flags {

		if !c.GlobalIsSet(f) {
			continue
		}

		sp := strings.Split(f, "-")

		if len(sp) < 2 {
			continue
		}

		if sp[0] != "firefox" {
			continue
		}

		optionName := flect.Pascalize(strings.Join(sp[1:], " "))
		var destVal interface{}

		// Find the corresponding flag
		for _, ff := range FirefoxGlobalFlags {
			if ff.GetName() == f {

				// Type switch on the flag type
				switch ff.(type) {

				case cli.StringFlag:
					destVal = c.GlobalString(f)

				}

			}
		}

		err := config.RegisterModuleOpt(mozilla.ConfigName,
			optionName, destVal)
		if err != nil {
			fflog.Fatal(err)
		}

	}

	return nil
}
