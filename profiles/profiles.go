// +build linux
//

package profiles

const (
	XDG_HOME = "XDG_CONFIG_HOME"
)

type ProfileManager interface {
	ListProfiles() []string
	GetProfiles() []*Profile
	GetDefaultProfile() Profile
}

type Profile struct {
	Id   string
	Name string
	Path string
}
