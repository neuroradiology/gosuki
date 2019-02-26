package config

import (
	"fmt"
	"gomark/logging"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/gobuffalo/flect"

	"github.com/fatih/structs"
)

var log = logging.GetLogger("CONF")

var C *GlobalConfig

const (
	ConfigFile = "config.toml"
)

type Config map[string]interface{}

type GlobalConfig struct {
	Firefox Config
	Chrome  Config
}

func RegisterGlobalVal(key string, val interface{}) error {
	log.Debugf("Registring global option %s = %v", key, val)
	//s := structs.New(C)

	return nil
}

func RegisterConf(module string, val interface{}) error {
	module = flect.Pascalize(module)
	log.Debug("registering config for module ", module)
	s := structs.New(C)

	// Store option in global config
	if module == "" {
		log.Debugf("Registring global conf  %v", val)

		// Store option in a config submodule
	} else {

		log.Debugf("Registering conf module <%s>  = %v", module, val)

		field, ok := s.FieldOk(module)

		if !ok {
			return fmt.Errorf("Module <%s> not defined in config", module)
		}

		err := field.Set(val)
		if err != nil {
			return err
		}

	}

	return nil
}

//func LoadConfigFile() error {

//}

func InitConfigFile() error {
	configFile, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}

	tomlEncoder := toml.NewEncoder(configFile)
	err = tomlEncoder.Encode(C)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigFile() (*GlobalConfig, error) {
	_, err := toml.DecodeFile(ConfigFile, C)

	return C, err
}

func init() {
	C = new(GlobalConfig)
}
