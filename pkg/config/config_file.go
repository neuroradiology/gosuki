// Copyright (c) 2023 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
// All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This file is part of GoSuki.
//
// GoSuki is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// GoSuki is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
package config

//TODO: load config path from cli flag/env var

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"

	"github.com/blob42/gosuki/internal/utils"
)

const (
	ConfigFileName = "config.toml"
	ConfigDirName  = "gosuki"
)

func getDefaultConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get config dir: %s", err)
	}
	if configDir == "" {
		return "", errors.New("could not get config dir")
	}

	configDir = path.Join(configDir, ConfigDirName)
	return configDir, nil
}

func getDefaultConfigPath() (string, error) {
	if configDir, err := getDefaultConfigDir(); err != nil {
		return "", err
	} else {
		return path.Join(configDir, ConfigFileName), nil
	}
}

func DefaultConfPath() string {
	configFile, err := getDefaultConfigPath()
	if err != nil {
		log.Fatal(err)
	}

	return configFile
}

func ConfigExists(path string) (bool, error) {
	return utils.CheckFileExists(path)
}

// Create a default toml config file
func createDefaultConfFile() error {
	var configDir string
	var err error

	if configDir, err = getDefaultConfigDir(); err != nil {
		return err
	}

	if err = os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config dir: %w", err)
	}

	configFilePath := path.Join(configDir, ConfigFileName)

	configFile, err := os.Create(configFilePath)
	if err != nil {
		return err
	}

	allConf := GetAll()

	tomlEncoder := toml.NewEncoder(configFile)
	tomlEncoder.Indent = ""
	err = tomlEncoder.Encode(&allConf)
	if err != nil {
		return err
	}

	err = configFile.Close()
	if err != nil {
		return err
	}

	fmt.Printf("config written to %s\n", configFilePath)
	return err
}

// creates a config file given a target path
func createConfFile(path string) error {
	configFile, err := os.Create(path)
	if err != nil {
		return err
	}

	allConf := GetAll()

	tomlEncoder := toml.NewEncoder(configFile)
	tomlEncoder.Indent = ""
	if err := tomlEncoder.Encode(&allConf); err != nil {
		configFile.Close()
		return err
	}

	if err := configFile.Close(); err != nil {
		return err
	}

	return nil
}

// Loads gosuki configuation into the global config
func LoadFromTomlFile(path string) error {
	buffer := make(Config)
	_, err := toml.DecodeFile(path, &buffer)
	if err != nil {
		return fmt.Errorf("loading config file %w", err)
	}

	//DEBUG:
	// fmt.Println("Mem Config Keys:")
	// for k, _ := range configs {
	// 	fmt.Printf("%#v\n", k)
	// }

	for k, val := range buffer {
		// send the conf to its own module
		if _, ok := configs[k]; !ok {
			// log.Debugf("creating module config [%s]", k)
			configs[k] = make(Config)
		}
		err = configs[k].MapFrom(val)
		if err != nil {
			return fmt.Errorf("parsing config <%s>: %w", k, err)
		}

	}

	//DEBUG:
	// log.Debugf("loaded config from %s\n", configFile)
	// log.Debugf("config file %#v\n", buffer)
	// log.Debugf("loaded config %#v", configs)

	return err
}
