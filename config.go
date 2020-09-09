package main

import (
	"git.sp4ke.com/sp4ke/gomark/config"
	"git.sp4ke.com/sp4ke/gomark/utils"
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
		err = config.LoadFromTomlFile()
		if err != nil {
			log.Fatal(err)
		}
	}

}
