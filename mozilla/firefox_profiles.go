package mozilla

import (
	"gomark/profiles"
	"gomark/utils"
	"path/filepath"
	"regexp"

	ini "gopkg.in/ini.v1"
)

const (
	ProfilesFile = "profiles.ini"
)

var (
	ConfigFolder  = ".mozilla/firefox"
	ReIniProfiles = regexp.MustCompile(`(?i)profile`)
)

type FFProfileManager struct {
	basePath     string
	profilesFile *ini.File
}

func (pm *FFProfileManager) GetProfiles() ([]*profiles.Profile, error) {
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

func (pm *FFProfileManager) ListProfiles() []string {
	sections := pm.profilesFile.SectionStrings()
	var result []string
	for _, s := range sections {
		if ReIniProfiles.MatchString(s) {
			result = append(result, s)
		}
	}

	return result
}

func NewFFProfileManager() (*FFProfileManager, error) {
	profiles, err := ini.Load(filepath.Join(ConfigFolder, ProfilesFile))

	if err != nil {
		return nil, err
	}

	pm := &FFProfileManager{
		basePath:     ConfigFolder,
		profilesFile: profiles,
	}

	return pm, nil
}

func init() {
	ConfigFolder = filepath.Join(utils.GetHomeDir(), ConfigFolder)

	// Check if base folder exists
	configFolderExists, err := utils.CheckDirExists(ConfigFolder)
	if !configFolderExists {
		fflog.Criticalf("The base firefox folder <%s> does not exist",
			ConfigFolder)
	}

	if err != nil {
		fflog.Critical(err)
	}

}
