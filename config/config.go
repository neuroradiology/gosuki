// TODO: save config back to file
// TODO: global config options should be automatically shown in cli global flags
package config

import (
	"os"

	"git.sp4ke.xyz/sp4ke/gomark/logging"

	"github.com/BurntSushi/toml"
)

var (
	log            = logging.GetLogger("CONF")
	ConfReadyHooks []func()
	configs        = make(map[string]Configurator)
)

const (
	ConfigFile       = "config.toml"
	GlobalConfigName = "global"
)

// A Configurator allows multiple packages and modules to set and access configs
// which can be mapped to any output backend (toml, cli flags, env variables ...)
type Configurator interface {
	Set(opt string, v interface{}) error
	Get(opt string) (interface{}, error)
	Dump() map[string]interface{}
	MapFrom(interface{}) error
}

// Used to store the global config
type Config map[string]interface{}

func (c Config) Set(opt string, v interface{}) error {
	c[opt] = v
	return nil
}

func (c Config) Get(opt string) (interface{}, error) {
	return c[opt], nil
}

func (c Config) Dump() map[string]interface{} {
	return c
}

func (c Config) MapFrom(src interface{}) error {
	// Not used here
	return nil
}

// Register a global option ie. under [global] in toml file
func RegisterGlobalOption(key string, val interface{}) {
	log.Debugf("Registring global option %s = %v", key, val)
	configs[GlobalConfigName].Set(key, val)
}

func RegisterModuleOpt(module string, opt string, val interface{}) error {
	log.Debugf("Setting option for module <%s>: %s = %s", module, opt, val)
	dest := configs[module]
	return dest.Set(opt, val)
}

// Get all configs as a map[string]interface{}
// FIX: only print exported fields, parse tags for hidden fields
func GetAll() Config {
	result := make(Config)
	for k, c := range configs {
		result[k] = c
	}
	return result
}


// Create a toml config file
func InitConfigFile() error {
	configFile, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}

	allConf := GetAll()

	tomlEncoder := toml.NewEncoder(configFile)
	err = tomlEncoder.Encode(&allConf)
	if err != nil {
		return err
	}

	return nil
}

func LoadFromTomlFile() error {
	dest := make(Config)
	_, err := toml.DecodeFile(ConfigFile, &dest)

	for k, val := range dest {
		// send the conf to its own module
		if _, ok := configs[k]; ok {
			configs[k].MapFrom(val)
		}
	}

	return err
}

// Hooks registered here will be executed after the config package has finished
// loading the conf
func RegisterConfReadyHooks(hooks ...func()) {
	ConfReadyHooks = append(ConfReadyHooks, hooks...)
}

// A call to this func will run all registered config hooks
func RunConfHooks() {
	log.Debug("running config hooks")
	for _, f := range ConfReadyHooks {
		f()
	}
}

// A configurator can set options available under it's own module scope
// or under the global scope. A configurator implements the Configurator interface
func RegisterConfigurator(name string, c Configurator) {
	log.Debugf("Registering configurator %s", name)
	configs[name] = c
}

func init() {
	// Initialize the global config
	configs[GlobalConfigName] = make(Config)
}
