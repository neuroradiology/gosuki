//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

// Configuration
package firefox

import (
	"fmt"

	"git.blob42.xyz/gosuki/gosuki/internal/config"
	"git.blob42.xyz/gosuki/gosuki/internal/database"
	"git.blob42.xyz/gosuki/gosuki/pkg/browsers/mozilla"
	"git.blob42.xyz/gosuki/gosuki/pkg/modules"
	"git.blob42.xyz/gosuki/gosuki/pkg/parsing"
	"git.blob42.xyz/gosuki/gosuki/pkg/profiles"
	"git.blob42.xyz/gosuki/gosuki/pkg/tree"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

const (
	//TODO: auto detect firefox base dir based on OS and installed flavors
	// FirefoxBaseDir = "$HOME/.mozilla/firefox"
	DefaultProfile   = "default"

	// Default flavour to use
	BrowserName = mozilla.FirefoxFlavour
)

var (

	// firefox global config state.  
	FFConfig = NewFirefoxConfig()

	ffProfileLoader = &profiles.INIProfileLoader{
		//BasePath to be set at runtime in init
		ProfilesFile: mozilla.ProfilesFile,
	}

	FirefoxProfileManager = mozilla.NewMozProfileManager(
		ffProfileLoader,
	)
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

func setBookmarkDir(fc *FirefoxConfig) {
	var err error

	// load the default profile from the one defined in the config
	var profile *profiles.Profile
	if profile, err = FirefoxProfileManager.GetProfileByName(BrowserName, fc.Profile); err != nil {
		log.Fatal(err)
	}

	bookmarkDir, err := profile.AbsolutePath()
	if err != nil {
		log.Fatal(err)
	}

	fc.BkDir = bookmarkDir
	log.Debugf("Using profile %s", bookmarkDir)
}

func NewFirefoxConfig() *FirefoxConfig {

	config := &FirefoxConfig{
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

	setBookmarkDir(config)

	return config
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

func init() {
	config.RegisterConfigurator(BrowserName, FFConfig)
}
