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

package database

import (
	"context"
	"math"
	"sync"
)

var (

	// lamport clock for this node
	Clock *LamportClock
)

type LamportClock struct {
	Value uint64
	mu    sync.RWMutex
}

func (c *LamportClock) Tick(peerClock uint64) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Value = uint64(math.Max(float64(c.Value), float64(peerClock)) + 1)
	// log.Trace("ticking clock", "prev", peerClock, "val", c.Value)
	return c.Value
}

func (c *LamportClock) LocalTick() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Trace("local ticking clock", "old", c.Value, "new", c.Value+1)
	c.Value += 1
	return c.Value
}

// GetDBClock returns lamport clock for this node's db (version column)
func (db *DB) GetDBClock(ctx context.Context) (*LamportClock, error) {
	var clock uint64

	if err := db.Handle.QueryRowContext(
		ctx,
		"select COALESCE(max(version),0) from gskbookmarks",
	).Scan(&clock); err != nil {
		return nil, err
	}

	return &LamportClock{clock, sync.RWMutex{}}, nil
}
