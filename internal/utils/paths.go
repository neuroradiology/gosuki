//
// Copyright ⓒ 2023 Chakib Ben Ziane <contact@blob42.xyz> and [`GoSuki` contributors]
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
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func GetDefaultDBPath() string {
	return "./"
}

func CheckDirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}

	return false, err
}

func CheckFileExists(file string) (bool, error) {
	info, err := os.Stat(file)
	if err == nil {
		if info.IsDir() {
			errMsg := fmt.Sprintf("'%s' is a directory", file)
			return false, errors.New(errMsg)
		}

		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func CheckWriteable(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		// dir exists, make sure we can write to it
		testfile := path.Join(dir, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", dir)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %w", err)
		}
		fi.Close()
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		// dir doesnt exist, check that we can create it
		return os.Mkdir(dir, 0775)
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %w, incorrect permissions", err)
	}

	return err
}

// ExpandPath expands a path with environment variables and tilde
// Symlinks are followed by default
func ExpandPath(paths ...string) (string, error) {
	var homedir string
	var err error
	if homedir, err = os.UserHomeDir(); err != nil {
		return "", err
	}
	path := os.ExpandEnv(filepath.Join(paths...))

	if path[0] == '~' {
		path = filepath.Join(homedir, path[1:])
	}
	return filepath.EvalSymlinks(path)
}
