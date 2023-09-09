package main

import (
	"git.blob42.xyz/gomark/gosuki/internal/config"
	"git.blob42.xyz/gomark/gosuki/internal/utils"
)

func initDefaultConfig() {
	//TODO: handle chrome
	println("Creating default config: config.toml")

	err := config.InitConfigFile()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	log.Debugf("gosuki init config")

	// Check if config file exists
	exists, err := utils.CheckFileExists(config.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		// Initialize default initConfig
		//NOTE: if custom flags are passed before config.toml exists, falg
		//options will not be saved to the initial config.toml, this means
		//command line flags have higher priority than config.toml
		initDefaultConfig()
	} else {
		err = config.LoadFromTomlFile()
		if err != nil {
			log.Fatal(err)
		}
	}

}
