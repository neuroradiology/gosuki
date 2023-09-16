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

// TODO: generalize this package to handle any mozilla based browser
package mozilla

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"git.blob42.xyz/gomark/gosuki/internal/logging"
	_debug "git.blob42.xyz/gomark/gosuki/pkg/profiles"

	"github.com/go-ini/ini"
)

// ProfileManager interface
type ProfileManager = _debug.ProfileManager
type INIProfileLoader = _debug.INIProfileLoader
type PathGetter = _debug.PathGetter

const (
	ProfilesFile = "profiles.ini"
)

var (
	log           = logging.GetLogger("mozilla")
	ReIniProfiles = regexp.MustCompile(`(?i)profile`)

	ErrProfilesIni      = errors.New("could not parse profiles.ini file")
	ErrNoDefaultProfile = errors.New("no default profile found")

	// Common default profiles for mozilla/firefox based browsers
	DefaultProfileNames = map[string]string{
		"firefox-esr": "default-esr",
	}
)

type MozProfileManager struct {
	BrowserName  string
	ConfigDir    string
	ProfilesFile *ini.File
	PathGetter   PathGetter
	ProfileManager
}

func (pm *MozProfileManager) loadProfile() error {

	log.Debugf("loading profile from <%s>", pm.PathGetter.GetPath())
	pFile, err := ini.Load(pm.PathGetter.GetPath())
	if err != nil {
		return err
	}

	pm.ProfilesFile = pFile
	return nil
}

func (pm *MozProfileManager) GetProfiles() ([]*_debug.Profile, error) {
    err := pm.loadProfile()
    if err != nil {
      return nil, err
    }

	sections := pm.ProfilesFile.Sections()
	var filtered []*ini.Section
	var result []*_debug.Profile
	for _, section := range sections {
		if ReIniProfiles.MatchString(section.Name()) {
			filtered = append(filtered, section)

			p := &_debug.Profile{
				Id: section.Name(),
			}

			err := section.MapTo(p)
			if err != nil {
				return nil, err
			}


			result = append(result, p)

		}
	}

	return result, nil
}

// GetProfilePath returns the absolute directory path to a mozilla profile.
func (pm *MozProfileManager) GetProfilePath(name string) (string, error) {
	log.Debugf("using config dir %s", pm.ConfigDir)
	p, err := pm.GetProfileByName(name)
	if err != nil {
		return "", err
	}
	rawPath := filepath.Join(pm.ConfigDir, p.Path)
	fullPath , err := filepath.EvalSymlinks(rawPath)

	return fullPath, err
	
	// Eval symlinks
}

func (pm *MozProfileManager) GetProfileByName(name string) (*_debug.Profile, error) {
	profs, err := pm.GetProfiles()
	if err != nil {
		return nil, err
	}

	for _, p := range profs {
		if p.Name == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("profile %s not found", name)
}

// TEST:
func (pm *MozProfileManager) GetDefaultProfile() (*_debug.Profile, error) {
	profs, err := pm.GetProfiles()
	if err != nil {
		return nil, err
	}

	defaultProfileName, ok := DefaultProfileNames[pm.BrowserName]
	if !ok {
		defaultProfileName = "default"
	}

	log.Debugf("looking for profile %s", defaultProfileName)
	for _, p := range profs {
		if p.Name == defaultProfileName {
			return p, nil
		}
	}

	return nil, ErrNoDefaultProfile
}

func (pm *MozProfileManager) ListProfiles() ([]string, error) {
	pm.loadProfile()
	sections := pm.ProfilesFile.SectionStrings()
	var result []string
	for _, s := range sections {
		if ReIniProfiles.MatchString(s) {
			result = append(result, s)
		}
	}

	if len(result) == 0 {
		return nil, ErrProfilesIni
	}

	return result, nil
}
