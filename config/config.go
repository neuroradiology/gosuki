package config

import (
	"gomark/logging"
	"os"

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

func RegisterGlobalOption(key string, val interface{}) {
	log.Debugf("Registring global option %s = %v", key, val)
	configs[GlobalConfigName].Set(key, val)
}

func RegisterModuleOpt(module string, opt string, val interface{}) {
	log.Debugf("adding option %s = %s", opt, val)
	dest := configs[module]
	dest.Set(opt, val)
}

// Get all configs as a map[string]interface{}
func GetAll() Config {
	result := make(Config)
	for k, c := range configs {
		result[k] = c
	}
	return result
}

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

// TODO: parse config back to each module conf
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

func RegisterConfReadyHooks(hooks ...func()) {
	for _, f := range hooks {
		ConfReadyHooks = append(ConfReadyHooks, f)
	}
}

func RunConfHooks() {
	for _, f := range ConfReadyHooks {
		f()
	}
}

func RegisterConfigurator(name string, c Configurator) {
	log.Debugf("Registering configurator %s", name)
	configs[name] = c
}

func init() {
	// Initialize the global config
	configs[GlobalConfigName] = make(Config)

}
