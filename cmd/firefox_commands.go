//TODO: add cli options to set/get options
package cmd

import (
	"fmt"
	"path/filepath"

	"git.sp4ke.com/sp4ke/gomark/logging"
	"git.sp4ke.com/sp4ke/gomark/mozilla"

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

func ffListProfiles(c *cli.Context) error {
	profs, err := mozilla.FirefoxProfileManager.GetProfiles()
	if err != nil {
		return err
	}

	for _, p := range profs {
		fmt.Printf("<%s>: %s\n", p.Name, filepath.Join(mozilla.GetBookmarkDir(), p.Path))
	}

	return nil
}

func ffCheckVFS(c *cli.Context) error {
	err := mozilla.CheckVFSLock()
	if err != nil {
		return err
	}

	return nil
}

func ffUnlockVFS(c *cli.Context) error {
	err := mozilla.UnlockPlaces()
	if err != nil {
		return err
	}

	return nil
}
