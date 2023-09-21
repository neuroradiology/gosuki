package mozilla

import (
	"git.blob42.xyz/gosuki/gosuki/pkg/profiles"
	"github.com/stretchr/testify/assert"
	"testing"
)

var OkProfile = &profiles.INIProfileLoader{
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

var BadProfile = &profiles.INIProfileLoader{
	BasePath:     "testdata",
	ProfilesFile: "profiles_bad.ini",
}

func TestGetProfiles(t *testing.T) {
	// fake browser flavour
	MozBrowsers["test"] = profiles.BrowserFlavour{"test", "testdata"}
	t.Run("OK", func(t *testing.T) {
		pm := &MozProfileManager{
			PathResolver: OkProfile,
		}

		profs, err := pm.GetProfiles("test")
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
	})
	t.Run("Bad", func(t *testing.T) {
		pm := &MozProfileManager{
			PathResolver: BadProfile,
		}

		_, err := pm.GetProfiles("test")
		if err != ErrProfilesIni || err == nil {
			t.Error("Expected error parsing bad profiles file")
		}
	})
}
