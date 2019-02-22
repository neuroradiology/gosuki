//TODO: Add cli options to choose firefox profile
package mozilla

import (
	"errors"
	"gomark/profiles"
	"gomark/utils"
	"path/filepath"
	"regexp"

	ini "gopkg.in/ini.v1"
)

type ProfileManager = profiles.ProfileManager
type ProfileGetter = profiles.ProfileGetter
type PathGetter = profiles.PathGetter

const (
	ProfilesFile = "profiles.ini"
)

var (
	ConfigFolder  = ".mozilla/firefox"
	ReIniProfiles = regexp.MustCompile(`(?i)profile`)

	firefoxProfile = &ProfileGetter{
		//BasePath to be set at runtime in init
		ProfilesFile: ProfilesFile,
	}

	FirefoxProfileManager = &FFProfileManager{
		pathGetter: firefoxProfile,
	}

	ErrProfilesIni      = errors.New("Could not parse Firefox profiles.ini file")
	ErrNoDefaultProfile = errors.New("No default profile found")
)

type FFProfileManager struct {
	profilesFile *ini.File
	pathGetter   PathGetter
	ProfileManager
}

func (pm *FFProfileManager) loadProfile() error {

	log.Debugf("loading profile from <%s>", pm.pathGetter.Get())
	pFile, err := ini.Load(pm.pathGetter.Get())
	if err != nil {
		return err
	}

	pm.profilesFile = pFile
	return nil
}

func (pm *FFProfileManager) GetProfiles() ([]*profiles.Profile, error) {
	pm.loadProfile()
	sections := pm.profilesFile.Sections()
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

func (pm *FFProfileManager) GetDefaultProfilePath() (string, error) {
	log.Debugf("using config dir %s", ConfigFolder)
	p, err := pm.GetDefaultProfile()
	if err != nil {
		return "", err
	}
	return filepath.Join(ConfigFolder, p.Path), nil
}

func (pm *FFProfileManager) GetDefaultProfile() (*profiles.Profile, error) {
	profs, err := pm.GetProfiles()
	if err != nil {
		return nil, err
	}

	for _, p := range profs {
		if p.Name == "default" {
			return p, nil
		}
	}

	return nil, ErrNoDefaultProfile
}

func (pm *FFProfileManager) ListProfiles() ([]string, error) {
	pm.loadProfile()
	sections := pm.profilesFile.SectionStrings()
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

func init() {
	ConfigFolder = filepath.Join(utils.GetHomeDir(), ConfigFolder)

	// Check if base folder exists
	configFolderExists, err := utils.CheckDirExists(ConfigFolder)
	if !configFolderExists {
		log.Criticalf("The base firefox folder <%s> does not exist",
			ConfigFolder)
	}

	if err != nil {
		log.Critical(err)
	}

	firefoxProfile.BasePath = ConfigFolder

	bookmarkDir, err := FirefoxProfileManager.GetDefaultProfilePath()
	if err != nil {
		log.Error(err)
	}

	SetBookmarkDir(bookmarkDir)

}
