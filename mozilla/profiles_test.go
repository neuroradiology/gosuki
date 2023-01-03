package mozilla

import (
	"testing"
)

var OkProfile = &INIProfileLoader{
	BasePath:     "testdata",
	ProfilesFile: "profiles_ok.ini",
}

var BadProfile = &INIProfileLoader{
	BasePath:     "testdata",
	ProfilesFile: "profiles_bad.ini",
}

func TestListProfiles(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		pm := &MozProfileManager{
			PathGetter: OkProfile,
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
		pm := &MozProfileManager{
			PathGetter: BadProfile,
		}

		_, err := pm.ListProfiles()

		if err != ErrProfilesIni || err == nil {
			t.Error("Expected error parsing bad profiles file")
		}
	})

}

func TestGetProfiles(t *testing.T) {
	pm := &MozProfileManager{
		PathGetter: OkProfile,
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
