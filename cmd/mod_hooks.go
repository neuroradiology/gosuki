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

// Modules can register custom hooks here that will plug into urfave *cli.App
// API. The hooks will be called in the same order as defined urfave's cli.
package cmd

import "github.com/urfave/cli/v2"

type Hook func(*cli.Context) error

// Map module id to list of *cli.App.Before hooks
var modCmdBeforeHooks = map[string]Hook{}

// Register a module hook to be run in *cli.App.Before
func RegBeforeHook(modID string, hook Hook) {
	if hook == nil {
		log.Fatalf("cannot register nil hook for <%s>", modID)
	}

	if _, ok := modCmdBeforeHooks[modID]; ok {
		log.Warnf("a hook was already registered for module <%s>", modID)
	}
	modCmdBeforeHooks[modID] = hook
}

// Return all registered Before hooks for module
func BeforeHook(modID string) Hook {
	if hook := modCmdBeforeHooks[modID]; hook != nil {
		return hook
	}
	return nil
}
