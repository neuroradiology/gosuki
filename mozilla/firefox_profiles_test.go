package mozilla

import (
	"testing"
)

var OkProfile = &ProfileGetter{
	BasePath:     "testdata",
	ProfilesFile: "profiles_ok.ini",
}

var BadProfile = &ProfileGetter{
	BasePath:     "testdata",
	ProfilesFile: "profiles_bad.ini",
}

func TestListProfiles(t *testing.T) {
	//_, filename, _, _ := runtime.Caller(0)
	//dir, err := filepath.Abs(filepath.Dir(filename))
	//if err != nil {
	//t.Error(err)
	//}
	//t.Error(dir)
	t.Run("OK", func(t *testing.T) {
		pm := &FFProfileManager{
			pathGetter: OkProfile,
		}

		t.Log("Listing profiles")
		profiles, err := pm.ListProfiles()
		if err != nil {
			t.Error(err)
		}

		for _, p := range profiles {
			t.Logf("found profiles: %s", p)
		}
		if profiles[0] != "Profile0" {
			t.Error("Expected Profile0")
		}
	})

	t.Run("Bad", func(t *testing.T) {
		pm := &FFProfileManager{
			pathGetter: BadProfile,
		}

		_, err := pm.ListProfiles()

		if err != ErrProfilesIni || err == nil {
			t.Error("Expected error parsing bad profiles file")
		}
	})

}

func TestGetProfiles(t *testing.T) {
	pm := &FFProfileManager{
		pathGetter: OkProfile,
	}

	profs, err := pm.GetProfiles()
	if err != nil {
		t.Error(err)
	}

	for _, p := range profs {
		t.Log(p)
	}

	if profs[0].Name != "default" {
		t.Error("Expected default profile in profiles.ini")
	}
}
