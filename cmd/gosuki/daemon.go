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
	"fmt"
	"os"

	"git.blob42.xyz/gosuki/gosuki/internal/api"
	db "git.blob42.xyz/gosuki/gosuki/internal/database"
	"git.blob42.xyz/gosuki/gosuki/internal/utils"
	"git.blob42.xyz/gosuki/gosuki/pkg/modules"
	"git.blob42.xyz/gosuki/gosuki/pkg/profiles"
	"git.blob42.xyz/gosuki/gosuki/pkg/watch"

	"git.blob42.xyz/gosuki/gosuki/pkg/manager"

	"github.com/urfave/cli/v2"
)


var startDaemonCmd = &cli.Command{
	Name:    "daemon",
	Aliases: []string{"d"},
	Usage:   "run browser watchers",
	// Category: "daemon"
	Action:  startDaemon,
}

// Runs the module by calling the setup 
func runModule(m *manager.Manager,
				c *cli.Context,
				browserMod modules.BrowserModule,
				p *profiles.Profile) (error) {
		var profileName string
		mod := browserMod.ModInfo()
		// Create context
		modContext := &modules.Context{
			Cli: c,
		}
		//Create a browser instance
		browser, ok := mod.New().(modules.BrowserModule)
		if !ok {
			return fmt.Errorf("module <%s> is not a BrowserModule", mod.ID)
		}
		log.Debugf("created browser instance <%s>", browser.Config().Name)

		// shutdown logic
		_, isShutdowner := browser.(modules.Shutdowner)
		if !isShutdowner {
			log.Warningf("browser <%s> does not implement modules.Shutdowner", browser.Config().Name)
		}


		if p != nil {
			bpm, ok := browser.(profiles.ProfileManager)
			if !ok {
				err := fmt.Errorf("<%s> does not implement profiles.ProfileManager",
				browser.Config().Name)
				log.Critical(err)
				return err
			}
			if err := bpm.UseProfile(*p); err != nil {
				log.Criticalf("could not use profile <%s>", p.Name)
				return err
			}
			profileName = p.Name
		}


		// calls the setup logic for each browser instance which
		// includes the browsers.Initializer and browsers.Loader interfaces
		//PERF:
		err := modules.Setup(browser, modContext, p)
		if err != nil {
			log.Errorf("setting up <%s> %v", browser.Config().Name, err)
			return err
		}

		runner, ok := browser.(watch.WatchRunner)
		if !ok {
			err =  fmt.Errorf("<%s> must implement watch.WatchRunner interface", browser.Config().Name)
			log.Critical(err)
			return err
		}

		w := runner.Watch()
		if w == nil {
			err = fmt.Errorf("<%s> must return a valid watch descriptor", browser.Config().Name)
			log.Critical(err)
			return err
		}
		log.Infof("adding watch runner <%s>", runner.Watch().ID)

		// create the worker name
		unitName := browser.Config().Name
		if len(profileName) > 0 {
			unitName = fmt.Sprintf("%s(%s)", unitName, profileName)
		}


		//BUG: last worker is the only instance that is run
		worker := watch.Worker(runner)

		m.AddUnit(worker, unitName)

		return nil
}

func startDaemon(c *cli.Context) error {
	defer utils.CleanFiles()
	manager := manager.NewManager()
	manager.ShutdownOn(os.Interrupt)

	api := api.NewApi()
	manager.AddUnit(api, "api")


	// Initialize sqlite database available in global `cacheDB` variable
	db.InitDB()

	registeredBrowsers := modules.GetBrowserModules()

	// instanciate all browsers
	for _, browserMod := range registeredBrowsers {
		mod := browserMod.ModInfo()

		//Create a temporary browser instance to check if it implements
		// the ProfileManager interface
		browser, ok := mod.New().(modules.BrowserModule)
		if !ok {
			log.Criticalf("TODO: module <%s> is not a BrowserModule", mod.ID)
		}

		// if the module is a profile manager and is watching all profiles
		// call runModule for each profile
		bpm, ok := browser.(profiles.ProfileManager)
		if ok {
			if bpm.WatchAllProfiles() {
				profs, err := bpm.GetProfiles()
				if err != nil {
					log.Critical("could not get profiles")
					continue
				}
				for _, p := range profs {
					log.Debugf("profile: <%s>", p.Name)
					err = runModule(manager, c, browserMod, p)
					if err != nil {
						log.Critical(err)
						continue
					}
				}
			} else {
				log.Debugf("profile manager <%s> not watching all profiles",
				browser.Config().Name)
				err := runModule(manager, c, browserMod, nil)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		} else {
			log.Warningf("module <%s> does not implement profiles.ProfileManager",
			browser.Config().Name)
			if err := runModule(manager, c, browserMod, nil); err != nil {
				log.Error(err)
				continue
			}
		}
	}

	go manager.Run()

	<-manager.Quit

	return nil
}
