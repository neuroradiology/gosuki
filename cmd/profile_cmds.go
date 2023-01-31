package cmd

import (
	"fmt"

	"git.sp4ke.xyz/sp4ke/gomark/modules"
	"git.sp4ke.xyz/sp4ke/gomark/profiles"
	"github.com/urfave/cli/v2"
)



var ProfileCmds = &cli.Command{
	Name: "profile",
	Usage: "profile commands",
	Subcommands: []*cli.Command{
		listProfilesCmd,
	},
}


//TODO: only enable commands when modules which implement profiles interfaces
// are available
var listProfilesCmd = &cli.Command{
	Name: "list",
	Usage: "list available profiles",
	Action: func(c *cli.Context) error {

	browsers := modules.GetBrowserModules()
	for _, br := range browsers {

		//Create a browser instance
		brmod, ok := br.ModInfo().New().(modules.BrowserModule)
		if !ok {
			log.Criticalf("module <%s> is not a BrowserModule", br.ModInfo().ID)
		}

		pm, isProfileManager := brmod.(profiles.ProfileManager)
		if !isProfileManager{
			log.Critical("not profile manager")
		}
		if isProfileManager {
			// handle default profile commands

			profs, err := pm.GetProfiles()
			if err != nil {
				return err
			}

			for _, p := range profs {
				fmt.Printf("%-10s \t %s\n", p.Name, pm.GetProfilePath(*p))
			}

			
		}
	}

		return nil
	},
}
