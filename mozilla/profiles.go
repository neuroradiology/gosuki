// TODO: generalize this package to handle any mozilla based browser
package mozilla

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"

	"git.blob42.xyz/gomark/gosuki/logging"
	"git.blob42.xyz/gomark/gosuki/profiles"

	"github.com/go-ini/ini"
)

// ProfileManager interface
type ProfileManager = profiles.ProfileManager
type INIProfileLoader = profiles.INIProfileLoader
type PathGetter = profiles.PathGetter

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

func (pm *MozProfileManager) GetProfiles() ([]*profiles.Profile, error) {
    err := pm.loadProfile()
    if err != nil {
      return nil, err
    }

	sections := pm.ProfilesFile.Sections()
	var filtered []*ini.Section
	var result []*profiles.Profile
	for _, section := range sections {
		if ReIniProfiles.MatchString(section.Name()) {
			filtered = append(filtered, section)

			p := &profiles.Profile{
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

// GetProfilePath returns the absolute path to a mozilla profile.
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

func (pm *MozProfileManager) GetProfileByName(name string) (*profiles.Profile, error) {
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
func (pm *MozProfileManager) GetDefaultProfile() (*profiles.Profile, error) {
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
