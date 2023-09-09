// Configuration
package firefox

import (
	"fmt"
	"os"
	"path/filepath"

	"git.blob42.xyz/gomark/gosuki/config"
	"git.blob42.xyz/gomark/gosuki/database"
	"git.blob42.xyz/gomark/gosuki/modules"
	"git.blob42.xyz/gomark/gosuki/mozilla"
	"git.blob42.xyz/gomark/gosuki/parsing"
	"git.blob42.xyz/gomark/gosuki/tree"
	"git.blob42.xyz/gomark/gosuki/utils"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/cli/v2"
)

const (
	BrowserName      = "firefox"
	FirefoxBaseDir = "$HOME/.mozilla/firefox"
	DefaultProfile   = "default"
)

var (

	// firefox global config state.  
	FFConfig = &FirefoxConfig{
		BrowserConfig: &modules.BrowserConfig{
			Name:         BrowserName,
			Type:         modules.TFirefox,
			BkDir:        "",
			BkFile:       mozilla.PlacesFile,
			NodeTree: &tree.Node{
				Name: mozilla.RootName,
				Parent: nil,
				Type:   tree.RootNode,
			},
			Stats:          &parsing.Stats{},
			UseFileWatcher: true,
			// NOTE: see parsing.Hook to add custom parsing logic for each
			// parsed node
			UseHooks:   []string{"notify-send"},
		},

		// Default data source name query options for `places.sqlite` db
		PlacesDSN: database.DsnOptions{
			"_journal_mode": "WAL",
		},

		// default profile to use, can be selected as cli argument
		Profile: DefaultProfile,

	}

	ffProfileLoader = &mozilla.INIProfileLoader{
		//BasePath to be set at runtime in init
		ProfilesFile: mozilla.ProfilesFile,
	}

	FirefoxProfileManager = &mozilla.MozProfileManager{
		BrowserName: BrowserName,
		PathGetter:  ffProfileLoader,
	}
)

// FirefoxConfig implements the Configurator interface
// which allows it to register and set field through the Configurator.
//
// It is also used alongside cli_flags.go to dynamically register cli flags
// that can change this config (struct fields) from command line at runtime.
//
// The struct schema defines the parameters to pass on to firefox that can be
// overriden by users. Options defined here will automatically supported in the
// config.toml file as well as the command line flags. New command line flags or
// config file options will only be accepted if they are defined here.
type FirefoxConfig struct {
	// Default data source name query options for `places.sqlite` db
    PlacesDSN        database.DsnOptions
    Profile          string

	modules.ProfilePrefs `toml:"profile_options"`

    //TEST: ignore this field in config.Configurator interface
	// Embed base browser config
    *modules.BrowserConfig `toml:"-"`
}

func (fc *FirefoxConfig) Set(opt string, v interface{}) error {
	// log.Debugf("setting option %s = %v", opt, v)
	s := structs.New(fc)
	f, ok := s.FieldOk(opt)
	if !ok {
		return fmt.Errorf("%s option not defined", opt)
	}

	return f.Set(v)
}

func (fc *FirefoxConfig) Get(opt string) (interface{}, error) {
	s := structs.New(fc)
	f, ok := s.FieldOk(opt)
	if !ok {
		return nil, fmt.Errorf("%s option not defined", opt)
	}

	return f.Value(), nil
}

func (fc *FirefoxConfig) Dump() map[string]interface{} {
	s := structs.New(fc)
	return s.Map()
}

func (fc *FirefoxConfig) String() string {
	s := structs.New(fc)
	return fmt.Sprintf("%v", s.Map())
}

func (fc *FirefoxConfig) MapFrom(src interface{}) error {
	return mapstructure.Decode(src, fc)
}

//REFACT: 
// Hook called when the config is ready
func initFirefoxConfig(c *cli.Context) error {
	log.Debugf("<firefox> initializing config")

	// The following code is executed before the cli context is ready
	// so we cannot use cli flags here

	pm := FirefoxProfileManager

	// expand environment variables in path
	pm.ConfigDir = filepath.Join(os.ExpandEnv(FirefoxBaseDir))

	// Check if base folder exists
	exists, err := utils.CheckDirExists(pm.ConfigDir)
	if !exists {
		log.Criticalf("the base firefox folder <%s> does not exist", pm.ConfigDir)
	}

	if err != nil {
		log.Fatal(err)
		return err
	}

	// The next part prepares the default profile using the profile manager
	ffProfileLoader.BasePath = pm.ConfigDir



	// use default profile
	// WIP: calling multiple profiles uses the following logic
	bookmarkDir, err := FirefoxProfileManager.GetProfilePath(FFConfig.Profile)
	if err != nil {
		log.Fatal(err)
	}

	// update bookmark dir in firefox config
	//TEST: verify that bookmark dir is set before browser is started
	FFConfig.BkDir = bookmarkDir
	log.Debugf("Using profile %s", bookmarkDir)
	return nil
}

func init() {
	config.RegisterConfigurator(BrowserName, FFConfig)

	//BUG: initFirefoxConfig is is called too early
	config.RegisterConfReadyHooks(initFirefoxConfig)
}
