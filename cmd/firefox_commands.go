package cmd

import (
	"fmt"
	"gomark/logging"
	"gomark/mozilla"
	"path/filepath"

	"github.com/urfave/cli"
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
	Subcommands: []cli.Command{
		ffUnlockVFSCmd,
		ffCheckVFSCmd,
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
	Subcommands: []cli.Command{
		ffListProfilesCmd,
	},
}

var FirefoxCmds = cli.Command{
	Name:    "firefox",
	Aliases: []string{"ff"},
	Usage:   "firefox related commands",
	Subcommands: []cli.Command{
		ffVFSCommands,
		ffProfilesCmds,
	},
	//Action:  unlockFirefox,
}

func ffListProfiles(c *cli.Context) {
	profs, err := mozilla.FirefoxProfileManager.GetProfiles()
	if err != nil {
		fflog.Error(err)
	}

	for _, p := range profs {
		fmt.Printf("<%s>: %s\n", p.Id, filepath.Join(mozilla.BookmarkDir, p.Path))
	}
}

func ffCheckVFS(c *cli.Context) {
	err := mozilla.CheckVFSLock()
	if err != nil {
		fflog.Error(err)
	}
}

func ffUnlockVFS(c *cli.Context) {
	err := mozilla.UnlockPlaces()
	if err != nil {
		fflog.Error(err)
	}
}
