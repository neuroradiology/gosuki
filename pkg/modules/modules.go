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

// Modules allow the extension of gosuki to handle other types of browsers or
// source of data that can be turned into bookmarks.
// 
// # Module Types
//
// 	1. Browsers MUST implement the [BrowserModule] interface.
// 	2. Simple modules MUST implement the [Module] interface.
package modules

import (
	"errors"

	"github.com/urfave/cli/v2"
)

var (
	registeredModules []Module
)

type Context struct {
	Cli *cli.Context
}

// Every new module needs to register as a Module using this interface
type Module interface {
	ModInfo() ModInfo
}

// browser modules need to implement Browser interface
type BrowserModule interface {
	Browser
	Module
}

// Information related to the browser module
type ModInfo struct {
	ID ModID // Id of this module

	// New returns a pointer to a new instance of a gosuki module.
	// Browser modules MUST implement this method.
	New func() Module
}

type ModID string



func verifyModule(module Module) error {
	var err error

	mod := module.ModInfo()
	if mod.ID == "" {
		err = errors.New("gosuki module ID is missing")
	}
	if mod.New == nil {
		err = errors.New("missing ModInfo.New")
	}
	if val := mod.New(); val == nil {
		err = errors.New("ModInfo.New must return a non-nil module instance")
	}

	return err
}

func RegisterModule(module Module) {
	// do not register browser modules here
	_, bMod := module.(BrowserModule)
	if bMod {
		panic("use RegisterBrowser for browser modules")
	}
	
	if err := verifyModule(module); err != nil {
		panic(err)
	}
	//TODO: Register by ID
	registeredModules = append(registeredModules, module)
}


// Returns a list of registerd browser modules
func GetModules() []Module {
	return registeredModules
}

