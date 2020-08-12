package cmd

import (
	"strings"

	"git.sp4ke.com/sp4ke/gomark/config"
	"git.sp4ke.com/sp4ke/gomark/mozilla"
	"git.sp4ke.com/sp4ke/gomark/utils"

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
	for _, f := range c.App.Flags {

		if utils.Ins(f.Names(), "help") ||
			utils.Ins(f.Names(), "version") {
			continue
		}

		if !f.IsSet() {
			continue
		}

		sp := strings.Split(f.Names()[0], "-")

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
			if ff.String() == f.String() {

				// Type switch on the flag type
				switch ff.(type) {

				case *cli.StringFlag:
					destVal = f.Names()[0]

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
