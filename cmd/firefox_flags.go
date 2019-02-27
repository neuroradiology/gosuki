package cmd

import (
	"strings"

	"github.com/gobuffalo/flect"
	cli "gopkg.in/urfave/cli.v1"
)

const (
	FirefoxDefaultProfileFlag = "firefox-default-profile"
)

var FirefoxGlobalFlags = []cli.Flag{
	cli.StringFlag{
		Name:  FirefoxDefaultProfileFlag,
		Usage: "Set the default firefox `PROFILE` to use",
	},
}

func GlobalFirefoxFlagsManager(c *cli.Context) error {
	flags := c.GlobalFlagNames()
	for _, f := range flags {
		sp := strings.Split(f, "-")
		fflog.Critical(sp)

		if len(sp) < 2 {
			continue
		}

		if sp[0] != "firefox" {
			continue
		}

		optionName := flect.Pascalize(strings.Join(sp[1:], ""))
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

	}

	return nil
}
