//
//  Copyright (c) 2024-2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
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

package logging

import (
	"bytes"
	"strings"
	"sync"
)

// TailBuffer is a struct that maintains a buffer of the last N lines written to it.
type TailBuffer struct {
	buf bytes.Buffer  // Internal buffer to store the data written to the TailBuffer.
	que []string      // Queue to keep track of the last N lines written.
	n   int           // Number of lines to keep in the queue.
	mu  *sync.RWMutex // Mutex for thread-safe operations.
}

func NewTailBuffer(n int) *TailBuffer {
	return &TailBuffer{
		n:  n,
		mu: &sync.RWMutex{},
	}
}

func (t *TailBuffer) Lines() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.que
}

func (t *TailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf.Write(p)
	lines := strings.Split(t.buf.String(), "\n")
	// fmt.Printf("%#v\n", lines)

	if len(lines) == 1 && lines[0] != "" {
		lines = []string{lines[0]}
	}

	for _, line := range lines {
		if line != "" && line != "\n" {
			t.que = append(t.que, line)
		}
		if len(t.que) > t.n {
			t.que = t.que[len(t.que)-t.n:]
		}
	}

	return len(p), nil
}
