//
// Copyright (c) 2023-2025 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
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
	"context"
	"errors"
	"fmt"

	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/pkg/profiles"
	"github.com/blob42/gosuki/pkg/watch"
	"github.com/urfave/cli/v3"
)

var (
	registeredModules []Module
	disabledMods      = map[ModID]bool{}
)

type Context struct {
	context.Context

	Cli *cli.Command

	IsTUI bool
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

// PreLoader is an interface for modules which is run only once when the module
// starts. It should have the same effect as  Watchable.Run().
// Run() is automatically called for watched events, Load() is called once
// before starting to watch events.
//
// PreLoader allows modules to do a first pass of bookmark loading logic before
// the watcher threads is spawned
type PreLoader interface {

	// PreLoad() will be called right after a browser is initialized
	PreLoad(*Context) error
}

// This type of PreLoader simply returns a list of bookmarks and has no knwoledge
// of the loading and syncrhonization primitives. Mostly useful for simple
// modules (see `modules.BookmarksImporter`).
type DumbPreLoader interface {
	PreLoad() ([]*gosuki.Bookmark, error)
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
// 2. [PreLoader].Load(): Initial pre loading of data before any runtime loop
func SetupModule(mod Module, c *Context) error {

	modID := mod.ModInfo().ID
	log.Info("setting up", "module", modID)

	initializer, okInit := mod.(Initializer)
	if okInit {
		log.Debug("custom init", "module", modID)
		if err := initializer.Init(c); err != nil {
			return fmt.Errorf("initialization error: %w", err)
		}
	}

	// handle PreLoader interface
	preloader, ok := mod.(PreLoader)
	if ok {
		log.Debug("preloading", "module", modID)
		err := preloader.PreLoad(c)
		if err != nil {
			return fmt.Errorf("preloading error <%s>: %v", modID, err)
		}
	}

	// handle DumbPreLoader interface
	dumbPreloader, ok := mod.(DumbPreLoader)
	if ok {
		log.Debugf("<%s> preloading", modID)
		if err := database.LoadBookmarks(dumbPreloader.PreLoad, string(modID)); err != nil {
			return fmt.Errorf("preloading error <%s>: %v", modID, err)
		}

		// store bookmarks
	}

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

	registeredModules = append(registeredModules, module)

	//TIP: custom handling of watcher types
	// switch module.(type) {
	// case watch.Poller:
	// 	fmt.Println("this is interval fetcher")
	//
	// case watch.WatchRunner:
	// 	fmt.Println("this is watch runner")
	// }
}

// Returns a list of registerd modules
func GetModules() []Module {
	var result []Module
	for _, mod := range registeredModules {
		if !disabledMods[mod.ModInfo().ID] {
			result = append(result, mod)
		}
	}
	return result
}

func Disable(id ModID) {
	disabledMods[id] = true
}

func Disabled(id ModID) bool {
	return disabledMods[id]
}
