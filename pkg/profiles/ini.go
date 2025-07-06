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

package profiles

import "path/filepath"

// PathResolver allows custom path resolution for profiles
type PathResolver interface {
	GetPath() string
	SetBaseDir(string)
}

type INIProfileLoader struct {
	// The absolute path to the directory where profiles.ini is located
	BasePath     string
	ProfilesFile string
}

func (pg *INIProfileLoader) GetPath() string {
	return filepath.Join(pg.BasePath, pg.ProfilesFile)
}

func (pg *INIProfileLoader) SetBaseDir(dir string) {
	pg.BasePath = dir
}
