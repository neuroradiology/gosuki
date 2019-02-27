package main

import (
	"gomark/config"
	"gomark/utils"
)

func InitDefaultConfig() {
	//TODO: handle chrome
	log.Debug("Creating default config on config.toml")

	err := config.InitConfigFile()
	if err != nil {
		log.Fatal(err)
	}
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
		err = config.LoadFromTomlFile()
		if err != nil {
			log.Fatal(err)
		}

	}

	// Execute config hooks
	config.RunConfHooks()
}
