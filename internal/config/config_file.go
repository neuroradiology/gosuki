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

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Create a toml config file
func InitConfigFile() error {
	configFile, err := os.Create(ConfigFile)
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

	log.Debugf("loaded firefox config %#v", configs["firefox"])

	return err
}

