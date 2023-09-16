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
