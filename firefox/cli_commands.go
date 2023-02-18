// TODO: add cli options to set/get options
// TODO: move browser module commands to their own module packag
package firefox

import (
	"fmt"

	"git.blob42.xyz/gomark/gosuki/cmd"
	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/mozilla"
	"git.blob42.xyz/gomark/gosuki/utils"

	"github.com/urfave/cli/v2"
)

var fflog = logging.GetLogger("FF")

var ffUnlockVFSCmd = cli.Command{
	Name:    "unlock",
	Aliases: []string{"u"},
	Action:  ffUnlockVFS,
}

var ffCheckVFSCmd = cli.Command{
	Name:    "check",
	Aliases: []string{"c"},
	Action:  ffCheckVFS,
}

var ffVFSCommands = cli.Command{
	Name:  "vfs",
	Usage: "VFS locking commands",
	Subcommands: []*cli.Command{
		&ffUnlockVFSCmd,
		&ffCheckVFSCmd,
	},
}

var ffListProfilesCmd = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Action:  ffListProfiles,
}

var ffProfilesCmds = cli.Command{
	Name:    "profiles",
	Aliases: []string{"p"},
	Usage:   "Profiles commands",
	Subcommands: []*cli.Command{
		&ffListProfilesCmd,
	},
}

var FirefoxCmds = &cli.Command{
	Name:    "firefox",
	Aliases: []string{"ff"},
	Usage:   "firefox related commands",
	Subcommands: []*cli.Command{
		&ffVFSCommands,
		&ffProfilesCmds,
	},
	//Action:  unlockFirefox,
}

func init() {
	cmd.RegisterModCommand(BrowserName, FirefoxCmds)
}

//TODO: #54 define interface for modules to handle and list profiles
//FIX: Remove since profile listing is implemented at the main module level
func ffListProfiles(_ *cli.Context) error {
	profs, err := FirefoxProfileManager.GetProfiles()
	if err != nil {
		return err
	}

	for _, p := range profs {
		fmt.Printf("%-10s \t %s\n", p.Name, utils.ExpandPath(FirefoxProfileManager.ConfigDir, p.Path))
	}

	return nil
}

func ffCheckVFS(_ *cli.Context) error {
	err := mozilla.CheckVFSLock("path to profile")
	if err != nil {
		return err
	}

	return nil
}

func ffUnlockVFS(_ *cli.Context) error {
	err := mozilla.UnlockPlaces("path to profile")
	if err != nil {
		return err
	}

	return nil
}
