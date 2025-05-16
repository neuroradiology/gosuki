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

// TODO: get runtime build/git info  see:
// https://github.com/lightningnetwork/lnd/blob/master/build/version.go#L66
// https://raw.githubusercontent.com/lightningnetwork/lnd/refs/heads/master/build/version.go
package utils

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// These constants define the application version and follow the semantic
// versioning 2.0.0 spec (http://semver.org/).
var (
	// Commit stores the current commit of this build, which includes the
	// most recent tag, the number of commits since that tag (if non-zero),
	// the commit hash, and a dirty marker. This should be set using the
	// -ldflags during compilation.
	Commit string

	// CommitHash stores the current commit hash of this build.
	CommitHash string

	// RawTags contains the raw set of build tags, separated by commas.
	RawTags string

	// GoVersion stores the go version that the executable was compiled
	// with.
	GoVersion string
)

const (
	// AppMajor defines the major version of this binary.
	AppMajor uint = 0

	// AppMinor defines the minor version of this binary.
	AppMinor uint = 1

	// AppPatch defines the application patch for this binary.
	AppPatch uint = 0
)

// Version returns the application version as a properly formed string per the
// semantic versioning 2.0.0 spec (http://semver.org/).
func Version() string {
	// Start with the major, minor, and patch versions.
	version := fmt.Sprintf("%d.%d.%d", AppMajor, AppMinor, AppPatch)
	if CommitHash != "" {
		version = fmt.Sprintf("%s-git:%s", version, CommitHash[:8])
	}

	return version
}

// Tags returns the list of build tags that were compiled into the executable.
func Tags() []string {
	if len(RawTags) == 0 {
		return nil
	}

	return strings.Split(RawTags, ",")
}

// Get build information from the runtime.
func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		GoVersion = info.GoVersion
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				CommitHash = setting.Value

			case "-tags":
				RawTags = setting.Value
			}
		}
	}
}
