//
//  Copyright (c) 2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package events

import (
	"github.com/blob42/gosuki/pkg/modules"
	"github.com/blob42/gosuki/pkg/watch"
)

// TUIBus is a channel for sending messages to the Text User Interface (TUI).
// Modules should use this channel to signal changes in their loading status.
var (
	//TODO: add debounce to avoid congestion
	TUIBus = make(chan any)
)

// StartedLoadingMsg represents a message indicating that a module has started loading.
// It contains the ID of the module and the total number of items to be loaded.
type StartedLoadingMsg struct {
	ID    modules.ModID
	Total uint
}

// ProgressUpdateMsg represents a message indicating progress in loading a module.
// It includes the module's ID, the current count of loaded items, and the total number of items to be loaded.
type ProgressUpdateMsg struct {
	ID           modules.ModID
	Instance     modules.Module
	CurrentCount uint
	Total        uint
	NewBk        bool // used for new boomkarks after full load is over
}

// LoadingCompleteMsg represents a message indicating that a module has finished loading.
// It contains the ID of the module that has completed loading.
type LoadingCompleteMsg struct {
	ID modules.ModID
}

// Started a [watch.Runner] instance
type RunnerStarted struct {
	watch.WatchRunner
}
