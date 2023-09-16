//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
// (https://github.com/blob42/gosuki/graphs/contributors).
//
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify it under the terms of
// the GNU Affero General Public License as published by the Free Software Foundation,
// either version 3 of the License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR
// PURPOSE.  See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along with
// gosuki.  If not, see <http://www.gnu.org/licenses/>.

// TODO: save config back to file
// TODO: global config options should be automatically shown in cli global flags
package config

import (
	"os"

	"git.blob42.xyz/gomark/gosuki/internal/logging"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"
)

type Hook func(c *cli.Context) error

var (
	log            = logging.GetLogger("CONF")
	ConfReadyHooks []Hook
	configs        = make(map[string]Configurator)
)

const (
	ConfigFile       = "config.toml"
	GlobalConfigName = "global"
)

// A Configurator allows multiple packages and modules to set and access configs
// which can be mapped to any output format (toml, cli flags, env variables ...)
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

// Setup cli flag for global options
func SetupGlobalFlags() []cli.Flag {
	log.Debugf("Setting up global flags")
	flags := []cli.Flag{}
	for k, v := range configs[GlobalConfigName].Dump() {
		log.Debugf("Registering global flag %s = %v", k, v)

		// Register the option as a cli flag
		switch val := v.(type) {
			case string:
				flags = append(flags, &cli.StringFlag{
					Category: "_",
					Name:  k,
					Value: val,
				})

			case int:
				flags = append(flags, &cli.IntFlag{
					Category: "_",
					Name: k,
					Value: val,
				})

			case bool:
				flags = append(flags, &cli.BoolFlag{
					Category: "_",
					Name: k,
					Value: val,
				})

			default:
				log.Fatalf("unsupported type for global option %s", k)
		}
	}

	return flags
}

func RegisterModuleOpt(module string, opt string, val interface{}) error {
	log.Debugf("Setting option for module <%s>: %s = %v", module, opt, val)
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
func RegisterConfReadyHooks(hooks ...Hook) {
	ConfReadyHooks = append(ConfReadyHooks, hooks...)
}

// A call to this func will run all registered config hooks
func RunConfHooks(c *cli.Context) {
	log.Debug("running config hooks")
	for _, f := range ConfReadyHooks {
		err := f(c)
		if err != nil {
		  log.Fatalf("error running config hook: %v", err)
		}
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
