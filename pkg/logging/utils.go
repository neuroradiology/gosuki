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

import "strings"

// Cleanup a path by returning the last n segments of the path only.
func trimCallerPath(path string, n int) string {
	// lovely borrowed from charmbracelet/log which was borrowed from zap
	// nb. To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.

	// Return the full path if n is 0.
	if n <= 0 {
		return path
	}

	// Find the last separator.
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}

	for i := 0; i < n-1; i++ {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(path[:idx], '/')
		if idx == -1 {
			return path
		}
	}

	return path[idx+1:]
}
