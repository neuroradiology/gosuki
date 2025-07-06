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

//go:build linux

package browsers

import (
	"github.com/blob42/gosuki/internal/utils"
	"github.com/blob42/gosuki/pkg/logging"
)

var log = logging.GetLogger("BROWSERS")

// expands to the full path to the base directory
// if the package is a snap, use the snap directory
// func (p BaseDir) Expand() (string, error) {
// 	return utils.ExpandPath(p.Dir)
// }

// base directory without normalization
func (b BrowserDef) BaseDir() string {
	if b.snapDir != "" && isSnap(b.snapDir) {
		return b.snapDir
	}
	return b.baseDir
}

// Expands to the full path of base directory
// If browser installed as snap package, expand to the snap base dir
func (b BrowserDef) ExpandBaseDir() (string, error) {
	if b.snapDir != "" && isSnap(b.snapDir) {
		return utils.ExpandPath(b.snapDir)
	}

	return utils.ExpandPath(b.baseDir)
}

// detects whether path is a snap directory
func isSnap(dir string) bool {
	if dir == "" {
		return false
	}

	normDir, err := utils.ExpandOnly(dir)
	if err != nil {
		log.Errorf("snap path: %s", err)
		return false
	}

	ok, err := utils.DirExists(normDir)
	if err != nil {
		log.Debugf("snap path: %s : %s", dir, err)
	}
	return ok
}
