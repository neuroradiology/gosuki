package main

import (
	"git.sp4ke.xyz/sp4ke/gomark/config"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
)

func initDefaultConfig() {
	//TODO: handle chrome
	log.Debug("Creating default config on config.toml")

	err := config.InitConfigFile()
	if err != nil {
		log.Fatal(err)
	}
}

// FIX: make config init manual from main package
// HACK: this section is called well before module options/config parameters are
// initialized
func initConfig() {
	log.Debugf("initializing config")

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
