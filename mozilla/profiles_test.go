package mozilla

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var OkProfile = &INIProfileLoader{
	BasePath:     "testdata",
	ProfilesFile: "profiles_ok.ini",
}

var okPaths = []string{
	"path.default",
	"path.profile1",
}

var okNames = []string{
	"default",
	"profile1",
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


	var pPaths []string 
	var pNames []string
	for _, p := range profs {
		pPaths = append(pPaths, p.Path)
		pNames = append(pNames, p.Name)

		//TEST: Test the absolute path
		
	}
	assert.ElementsMatch(t, okPaths, pPaths)
	assert.ElementsMatch(t, okNames, pNames)

	if profs[0].Name != "default" {
		t.Error("Expected default profile in profiles.ini")
	}
}
