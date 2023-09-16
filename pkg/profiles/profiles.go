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


// go:build linux


const (
	XDGHome = "XDG_CONFIG_HOME"
)

// ProfileManager is any module that can detect or list profiles, usually a browser module. 
type ProfileManager interface {

	// Get all profile details
	GetProfiles() ([]*Profile, error)

	// Returns the default profile if no profile is selected
	GetDefaultProfile() (*Profile, error)

	// Return that absolute path to a profile and follow symlinks
	GetProfilePath(Profile) (string)

	// If should watch all profiles
	WatchAllProfiles() bool

	// Notifies the module to use a custom profile
	UseProfile(p Profile) error
}

type Profile struct {
	// Unique identifier for the profile
	Id   string

	// Name of the profile
	Name string

	// relative path to profile
	Path string
}

type PathGetter interface {
	GetPath() string
}
