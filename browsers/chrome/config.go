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

package chrome

import (
	"git.blob42.xyz/gosuki/gosuki/pkg/modules"
	"git.blob42.xyz/gosuki/gosuki/pkg/parsing"
	"git.blob42.xyz/gosuki/gosuki/pkg/tree"
)

const (
	BrowserName    = "chrome"
	ChromeBaseDir  = "$HOME/.config/google-chrome"
	DefaultProfile = "Default"
	RootNodeName   = "ROOT"
)

type ChromeConfig struct {
	Profile                string
	*modules.BrowserConfig `toml:"-"`
	modules.ProfilePrefs   `toml:"profile_options"`
}

var (
	ChromeCfg = &ChromeConfig{
		Profile: DefaultProfile,
		BrowserConfig: &modules.BrowserConfig{
			Name:   BrowserName,
			Type:   modules.TChrome,
			BkDir:  "$HOME/.config/google-chrome/Default",
			BkFile: "Bookmarks",
			NodeTree: &tree.Node{
				Name:   RootNodeName,
				Parent: nil,
				Type:   tree.RootNode,
			},
			Stats:          &parsing.Stats{},
			UseFileWatcher: true,
			UseHooks:       []string{"tags_from_name", "notify-send"},
		},
		//TODO: profile
	}
)
