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

// Package hooks permits to register custom hooks that will be called during the parsing
// process of a bookmark file. Hooks can be used to extract tags, commands or other custom
// data from a bookmark title or description.
//
// They can effectively be used as a command line interface to the host system
// through the browser builtin Ctrl+D bookmark feature.
//
// TODO: document types of hooks
package hooks

import (
	"reflect"
	"sort"

	"github.com/blob42/gosuki"
	"github.com/blob42/gosuki/pkg/tree"
)

type Hookable interface {
	*gosuki.Bookmark | *tree.Node
}

// A Hook is a function that takes a *Bookmark or *Node and runs an arbitrary process.
// Hooks are called during the loading or real time detection of bookmarks.
//
// For example the TAG extraction process is handled by the ParseXTags hooks.
//
// Hooks can also be used handle call custom user commands and messages found in the various fields of a bookmark.
type Hook[T Hookable] struct {
	// Unique name of the hook
	name string

	// Function to call on a node/bookmark
	Func func(T) error

	priority uint // hook order. Highest == 0
}

func (h Hook[T]) Name() string {
	return h.name
}

func SortByPriority(hooks []NamedHook) {
	sort.Slice(hooks, func(i, j int) bool {
		vi := reflect.ValueOf(hooks[i])
		vj := reflect.ValueOf(hooks[j])

		if vi.Kind() != reflect.Struct || vj.Kind() != reflect.Struct {
			panic("expected struct")
		}

		pi := vi.FieldByName("priority")
		pj := vj.FieldByName("priority")

		if !pi.IsValid() || !pj.IsValid() {
			panic("missing priority field")
		}

		return pi.Uint() < pj.Uint()
	})
}

// Browser who implement this interface will be able to register custom
// hooks which are called during the main Run() to handle commands and
// messages found in tags and parsed data from browsers
type HookRunner interface {

	// Calls all registered hooks on a node
	CallHooks(any) error
}
