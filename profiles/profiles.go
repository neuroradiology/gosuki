// Package profiles ...
package profiles


// go:build linux


const (
	XDG_HOME = "XDG_CONFIG_HOME"
)

// ProfileManager is any module that can detect or list profiles, usually a browser module. 
type ProfileManager interface {

	// Get all profile details
	GetProfiles() ([]*Profile, error)

	// Returns the default profile if no profile is selected
	GetDefaultProfile() (*Profile, error)

	// Return that absolute path to a profile and follow symlinks
	GetProfilePath(Profile) (string)

	// If should watch all profiles
	WatchAllProfiles() bool

	// Use custom profile
	UseProfile(p Profile) error
}

type Profile struct {
	// Unique identifier for the profile
	Id   string

	// Name of the profile
	Name string

	// relative path to profile
	Path string
}

type PathGetter interface {
	GetPath() string
}
