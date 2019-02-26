package main

import (
	"gomark/config"
	"gomark/mozilla"

	"github.com/fatih/structs"
)

var FirefoxDefaultConfig = mozilla.FirefoxConfig{

	// default profile to use
	DefaultProfile: "default",

	WatchAllProfiles: false,
}

func InitDefaultConfig() {
	log.Debug("Creating default config on config.toml")
	s := structs.New(FirefoxDefaultConfig)

	// Export default firefox config
	err := config.RegisterConf("firefox", s.Map())
	if err != nil {
		log.Fatal(err)
	}

	// Set default values for firefox module
	dest := structs.New(mozilla.Config)
	for _, f := range s.Fields() {
		if f.IsExported() {
			destF := dest.Field(f.Name())
			destF.Set(f.Value())
		}
	}

	err = config.InitConfigFile()
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
	dest := structs.New(mozilla.Config)
	for k, v := range c.Firefox {
		f := dest.Field(k)
		if f != nil {

			err := f.Set(v)
			if err != nil {
				log.Error(err)
				continue
			}

		}
	}

	log.Warningf("%#v", mozilla.Config)

}
