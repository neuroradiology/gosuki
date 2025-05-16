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

package hooks

// Global available hooks for browsers to use

import (
	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/pkg/parsing"
	"github.com/blob42/gosuki/pkg/tree"
)

type HookMap map[string]interface{}

type WithName interface {
	Name() string
}

var Predefined = HookMap{
	"node_tags_from_name": Hook[*tree.Node]{
		name: "node_tags_from_name",
		Func: parsing.ParseNodeTags,
	},
	"bk_tags_from_name": Hook[*gosuki.Bookmark]{
		name: "bk_tags_from_name",
		Func: parsing.ParseBkTags,
	},
}

func regHook(hooks ...WithName) {
	for _, hook := range hooks {
		Predefined[hook.Name()] = hook
	}
}
