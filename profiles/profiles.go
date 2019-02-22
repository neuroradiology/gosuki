// +build linux
//

package profiles

import "path/filepath"

const (
	XDG_HOME = "XDG_CONFIG_HOME"
)

type ProfileManager interface {
	ListProfiles() ([]string, error)
	GetProfiles() ([]*Profile, error)
	GetDefaultProfile() (*Profile, error)
}

type Profile struct {
	Id   string
	Name string
	Path string
}

func (p *Profile) GetPath() string {
	return p.Path
}

type PathGetter interface {
	Get() string
}

type ProfileGetter struct {
	BasePath     string
	ProfilesFile string
}

func (pg *ProfileGetter) Get() string {
	return filepath.Join(pg.BasePath, pg.ProfilesFile)
}
