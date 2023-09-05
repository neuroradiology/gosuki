package chrome

import (
	"git.blob42.xyz/gomark/gosuki/modules"
	"git.blob42.xyz/gomark/gosuki/parsing"
	"git.blob42.xyz/gomark/gosuki/tree"
)


const (
	BrowserName    = "chrome"
	ChromeBaseDir  = "$HOME/.config/google-chrome"
	DefaultProfile = "Default"
)

type ChromeConfig struct {
	Profile                string
	*modules.BrowserConfig `toml:"-"`
	modules.ProfilePrefs `toml:"profile_options"`
}

var (
	ChromeCfg = &ChromeConfig{
		Profile:       DefaultProfile,
		BrowserConfig: &modules.BrowserConfig{
			Name:           BrowserName,
			Type:           modules.TChrome,
			BkDir:          "",
			BkFile:         "Bookmarks",
			//TODO: build watched paths
			WatchedPaths:   []string{"/home/spike/.config/google-chrome-unstable/Default/Bookmarks"},
			NodeTree: &tree.Node{
				Name:   "ROOT",
				Parent: nil,
				Type: tree.RootNode,
			},
			Stats: &parsing.Stats{},
			UseFileWatcher: true,
		},
		//TODO: profile
	}
)
