package config

import (
	"gomark/logging"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
)

var (
	log            = logging.GetLogger("CONF")
	ConfReadyHooks []func()
	GlobalConfig   = Config{}
	C              = GlobalConfig
)

const (
	ConfigFile = "config.toml"
)

type Config map[string]interface{}

// Map a config to destination struct
func (c Config) MapTo(module string, dest interface{}) error {
	return mapstructure.Decode(c[module], dest)
}

func RegisterGlobalOption(key string, val interface{}) {
	log.Debugf("Registring global option %s = %v", key, val)
	C[key] = val
}

func RegisterBrowserConf(module string, val interface{}) {

	// Use global conf func instead
	if module == "" {
		return
	}

	// Store option in a config submodule
	log.Debugf("Registering conf module <%s>  = %v", module, val)
	C[module] = val
}

func InitConfigFile() error {
	configFile, err := os.Create(ConfigFile)
	if err != nil {
		return err
	}

	tomlEncoder := toml.NewEncoder(configFile)
	err = tomlEncoder.Encode(&C)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigFile() (Config, error) {
	_, err := toml.DecodeFile(ConfigFile, &C)

	return C, err
}

// Copies a src struct to dest struct
func MapConfStruct(src interface{}, dst interface{}) {
	s := structs.New(src)
	d := structs.New(dst)
	for _, f := range s.Fields() {
		if f.IsExported() {
			dF := d.Field(f.Name())
			dF.Set(f.Value())
		}
	}
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
