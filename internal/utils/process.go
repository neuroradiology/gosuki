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

package utils

import (
	"os"
	"path/filepath"

	psutil "github.com/shirou/gopsutil/process"
)

func FileProcessUsers(path string) (map[int32]*psutil.Process, error) {
	fusers := make(map[int32]*psutil.Process)

	processes, err := psutil.Processes()
	if err != nil &&
		err != os.ErrPermission {
		return nil, err
	}

	// Eval symlinks
	relPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}

	//log.Debugf("checking against path: %s", relPath)
	for _, p := range processes {

		files, err := p.OpenFiles()

		//TEST: use os.IsNotExist to test the path error
		if err != nil && os.IsNotExist(err) {
			continue
		}

		// Check if path in files

		for _, f := range files {
			//log.Debug(f)
			if f.Path == relPath {
				fusers[p.Pid] = p
			}
		}

	}

	return fusers, nil
}
