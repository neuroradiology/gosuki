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

// TODO: auto detect vfs lock then switch or not to watch&copy places
package mozilla

import (
	"errors"
	"fmt"
	"path"

	"github.com/blob42/gosuki/internal/utils"
)

const (
	// This option disables the VFS lock on firefox
	// Sqlite allows file locking of the database using the local file system VFS.
	// Previous versions of FF allowed external processes to access the file.
	//
	// Since firefox v(63) this has changed. When initializing the database FF checks
	// the preference option `storage.multiProcessAccess.enabled` which is not
	// documented officially.
	//
	// Source code:
	//- https://dxr.mozilla.org/mozilla-central/source/storage/TelemetryVFS.cpp#884
	//- https://dxr.mozilla.org/mozilla-central/source/storage/mozStorageService.cpp#377
	//- Change on github: https://github.com/mozilla/gecko-dev/commit/a543f35d4be483b19446304f52e4781d7a4a0a2f
	PrefMultiProcessAccess = "storage.multiProcessAccess.enabled"
)

var (
	ErrMultiProcessAlreadyEnabled = errors.New("multiProcessAccess already enabled")
)

// TODO!:
func CheckVFSLock(bkDir string) error {
	log.Debugf("TODO: checking VFS for <%s>", bkDir)
	return nil
}

func UnlockPlaces(bkDir string) error {
	log.Debugf("Unlocking VFS <%s>", path.Join(bkDir, PrefsFile))

	prefsPath := path.Join(bkDir, PrefsFile)

	// Find if multiProcessAccess option is defined

	pref, err := GetPrefBool(prefsPath, PrefMultiProcessAccess)
	if err != nil && err != ErrPrefNotFound {
		return err
	}

	// If pref already defined and true raise an error
	if pref {
		log.Warnf("pref <%s> already defined as <%v>",
			PrefMultiProcessAccess, pref)
		return nil

		// Set the preference
	} else {

		// Checking if firefox is running
		// TODO: #multiprocess add CLI to unlock places.sqlite
		pusers, err := utils.FileProcessUsers(path.Join(bkDir, PlacesFile))
		if err != nil {
			log.Error(err)
		}

		for pid, p := range pusers {
			pname, err := p.Name()
			if err != nil {
				log.Error(err)
			}
			return fmt.Errorf("multiprocess not enabled and %s(%d) is running. Close firefox and disable VFS lock", pname, pid)
		}
		// End testing

		// enable multi process access in firefox
		err = SetPrefBool(prefsPath,
			PrefMultiProcessAccess,
			true)

		if err != nil {
			return err
		}

	}

	return nil

}
