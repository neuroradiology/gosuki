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

package parsing

import (
	"time"
)

// This interface should be implemented by modules who can expose counters on
// numbers of available bookmarks
type Counter interface {

	// Current number of loaded bookmarks
	URLCount() uint
	NodeCount() uint

	// Progress in domain [0,1]
	Progress() float64

	// Reset all internal counters
	ResetCount()

	// Set total bookmarks to be loaded
	SetTotal(uint)

	// Total bookmark count that will be loaded
	Total() uint

	AddTotal(uint)

	SetLastWatchRuntime(time.Duration)

	SetLastTreeParseRuntime(time.Duration)
	LastFullTreeParseRT() time.Duration

	IncNodeCount()
	IncURLCount()
	SetURLCount(uint)
}

type BrowserCounter struct {
	lastFullTreeParseTime time.Duration
	lastWatchRunTime      time.Duration
	lastNodeCount         uint
	lastURLCount          uint
	currentNodeCount      uint
	currentURLCount       uint

	// Total URL count to be loaded
	totalURLCount uint
}

// AddTotal implements Counter.
func (c *BrowserCounter) AddTotal(n uint) {
	c.totalURLCount += n
}

// NodeCount implements Counter.
func (c *BrowserCounter) NodeCount() uint {
	return c.currentNodeCount
}

// LastFullTreeParseRT implements Counter.
func (c *BrowserCounter) LastFullTreeParseRT() time.Duration {
	return c.lastFullTreeParseTime
}

// IncUrlCount implements Counter.
func (c *BrowserCounter) IncURLCount() {
	c.currentURLCount++
}

// SetUrlCount implements Counter.
func (c *BrowserCounter) SetURLCount(count uint) {
	c.currentURLCount = count
}

// IncNodeCount implements Counter.
func (c *BrowserCounter) IncNodeCount() {
	c.currentNodeCount++
}

// SetLastTreeParseRuntime implements Counter.
func (c *BrowserCounter) SetLastTreeParseRuntime(d time.Duration) {
	c.lastFullTreeParseTime = d
}

// SetLastWatchTime implements Counter.
func (c *BrowserCounter) SetLastWatchRuntime(d time.Duration) {
	c.lastWatchRunTime = d
}

// URLCount implements Counter.
func (c *BrowserCounter) URLCount() uint {
	return c.currentURLCount
}

// Total implements Counter.
func (c *BrowserCounter) Total() uint {
	return c.totalURLCount
}

func (c *BrowserCounter) SetTotal(count uint) {
	c.totalURLCount = count
}

func (c *BrowserCounter) ResetCount() {
	c.lastURLCount = c.currentURLCount
	c.lastNodeCount = c.currentNodeCount
	c.currentNodeCount = 0
	c.currentURLCount = 0
	c.totalURLCount = 0
}

func (c *BrowserCounter) Progress() float64 {
	return float64(c.currentURLCount) / float64(c.totalURLCount)
}

// var _ Counter = (*BrowserCounter)(nil)
