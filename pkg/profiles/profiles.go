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

// Package profiles ...
package profiles

import (
	"github.com/blob42/gosuki/internal/utils"
	"github.com/blob42/gosuki/pkg/logging"
)

var log = logging.GetLogger("profiles")

// ProfileManager is any module that can detect or list profiles, usually a browser module.
// One profile manager should be created for each browser flavour.
type ProfileManager interface {

	// Returns all profiles for a given flavour
	GetProfiles(flavour string) ([]*Profile, error)

	// If should watch all profiles
	WatchAllProfiles() bool

	// Notifies the module to use a custom profile and flavour
	UseProfile(p *Profile, f *Flavour) error

	// Get current active profile
	GetProfile() *Profile

	// Returns all flavours supported by this module
	ListFlavours() []Flavour

	// Get current active flavour
	GetCurFlavour() *Flavour
}

// Returns flavour of browser given browser BaseDir
// TEST:
func GetFlavour(pm ProfileManager, baseDir string) string {
	flavours := pm.ListFlavours()
	for _, f := range flavours {
		if f.BaseDir == baseDir {
			return f.Name
		}
	}

	return ""
}

type Profile struct {
	// Unique identifier for the profile
	ID string

	// Name of the profile
	// This is usually the name of the directory where the profile is stored
	Name string

	// relative path to profile from base dir
	Path string

	// Base dir of the profile
	BaseDir string
}

func (p Profile) AbsolutePath() (string, error) {
	return utils.ExpandPath(p.BaseDir, p.Path)
}

// The Flavour struct stores the name of the browser and the base
// directory where the profiles are stored.
// Example flavours: chrome-stable, chrome-unstable, firefox, firefox-esr, librewolf, etc.
type Flavour struct {
	Name    string
	BaseDir string
}

// Detect if the browser is installed. Returns true if the path exists
func (b Flavour) Detect() bool {
	var dir string
	var err error
	if dir, err = utils.ExpandPath(b.BaseDir); err != nil {
		log.Warnf("could not expand path <%s>: %s", b.BaseDir, err)
		return false
	} else if ok, err := utils.DirExists(dir); err != nil || !ok {
		log.Warnf("could not find browser <%s> at <%s>: %v", b.Name, dir, err)
		return false
	}

	return true
}
