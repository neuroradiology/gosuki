package main

import (
	"gomark/config"
	"gomark/database"
	"gomark/mozilla"
	"gomark/utils"
)

// Config names
const (
	FirefoxConf = "firefox"
	ChromeConf  = "chrome"
)

var FirefoxDefaultConfig = mozilla.FirefoxConfig{

	// Default data source name query options for `places.sqlite` db
	PlacesDSN: database.DsnOptions{
		"_journal_mode": "WAL",
	},

	// default profile to use
	DefaultProfile: "default",

	WatchAllProfiles: false,
}

func InitDefaultConfig() {
	//TODO: handle chrome
	log.Debug("Creating default config on config.toml")

	// Export default firefox config
	config.RegisterBrowserConf(FirefoxConf, FirefoxDefaultConfig)

	// Set default values for firefox module
	config.MapConfStruct(FirefoxDefaultConfig, mozilla.Config)

	err := config.InitConfigFile()
	if err != nil {
		log.Fatal(err)
	}
}

// Loads config from config file and shares config with browser modules
func LoadConfig() {
	log.Debug("Loading config.toml")
	c, err := config.LoadConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	// Sync to firefox module config
	err = c.MapTo(FirefoxConf, mozilla.Config)

	if err != nil {
		log.Fatal(err)
	}

	log.Warningf("%#v", mozilla.Config)
}

func init() {

	// Check if config file exists
	exists, err := utils.CheckFileExists(config.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		// Initialize default config
		InitDefaultConfig()
	} else {
		//TODO: maybe no need to preload if we can preparse options with altsrc
		LoadConfig()
	}

	// Execute config hooks
	config.RunConfHooks()
}
