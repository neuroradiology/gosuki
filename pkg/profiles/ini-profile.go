package profiles

import "path/filepath"

type INIProfileLoader struct {
	// The absolute path to the directory where profiles.ini is located
	BasePath     string
	ProfilesFile string
}

func (pg *INIProfileLoader) GetPath() string {
	return filepath.Join(pg.BasePath, pg.ProfilesFile)
}
