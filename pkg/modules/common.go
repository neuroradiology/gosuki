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
//  1. Browsers MUST implement the [BrowserModule] interface.
//  2. Simple modules MUST implement the [Module] interface.
package modules

import (
	"errors"
	"fmt"

	"github.com/blob42/gosuki/pkg/profiles"
	"github.com/blob42/gosuki/pkg/watch"
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

// Information related to the browser module
type ModInfo struct {
	ID ModID // Id of this module

	// New returns a pointer to a new instance of a gosuki module.
	// Browser modules MUST implement this method.
	New func() Module
}

type ModID string

// Initialize the module before any data loading or callbacks
// If a module wants to do any preparation and prepare custom state before Loader.Load()
// is called and before any Watchable.Run() or other callbacks are executed.
type Initializer interface {

	// Init() is the first method called after a browser instance is created
	// and registered.
	// A pointer to
	// Return ok, error
	Init(*Context) error
}

// Loader is an interface for modules which is run only once when the module
// starts. It should have the same effect as  Watchable.Run().
// Run() is automatically called for watched events, Load() is called once
// before starting to watch events.
//
// Loader allows modules to do a first pass of Run() logic before the watcher
// threads is spawned
type Loader interface {

	// Load() will be called right after a browser is initialized
	Load() error
}

// ProfileInitializer is similar to Initializer but is called with a profile.
// This is useful for modules that need to do some custom initialization for a
// specific profile.
type ProfileInitializer interface {
	Init(*Context, *profiles.Profile) error
}

// Modules which implement this interface need to handle all shuttind down and
// cleanup logic in the defined methods. This is usually called at the end of
// the module instance lifetime
type Shutdowner interface {
	watch.Shutdowner
}

// SetupModule() is called for each registered module. The following methods, if
// implemented, are called in order:
//
// 1. [Initializer].Init(): state initialization
//
// 2. [Loader].Load(): Initial pre loading of data before any runtime loop
// TODO!:
func SetupModule(mod Module, c *Context) error {

	modID := mod.ModInfo().ID
	log.Infof("setting up module <%s>", modID)

	initializer, okInit := mod.(Initializer)
	if okInit {
		log.Debugf("<%s> custom init", modID)
		if err := initializer.Init(c); err != nil {
			return fmt.Errorf("<%s> initialization error: %w", modID, err)
		}
	}
	//TODO: ProfileInitializer

	return nil
}

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
	if _, bMod := module.(BrowserModule); bMod {
		log.Fatal("use RegisterBrowser for browser modules")
	}

	if err := verifyModule(module); err != nil {
		panic(err)
	}

	//TODO: Register by ID
	registeredModules = append(registeredModules, module)

	//WIP:
	switch module.(type) {
	case watch.IntervalFetcher:
		fmt.Println("this is interval fetcher")

	case watch.WatchRunner:
		fmt.Println("this is watch runner")
	}
}

// Returns a list of registerd browser modules
func GetModules() []Module {
	return registeredModules
}
