package mozilla

import (
	"os"
	"testing"
)

func TestNewProfileManager(t *testing.T) {
	InitialConfigFolder := ConfigFolder
	ConfigFolder = "toto"
	_, err := NewFFProfileManager()

	if !os.IsNotExist(err) {
		t.Error(err)
	}

	ConfigFolder = InitialConfigFolder
}

func TestListProfiles(t *testing.T) {
	pm, _ := NewFFProfileManager()

	t.Log("Listing profiles")
	profiles := pm.ListProfiles()
	for _, p := range pm.ListProfiles() {
		t.Logf("found profiles: %s", p)
	}
	if profiles[0] != "Profile0" {
		t.Error("Expected at least Profile0")
	}
}

func TestGetProfiles(t *testing.T) {
	pm, _ := NewFFProfileManager()
	profs, err := pm.GetProfiles()
	if err != nil {
		t.Error(err)
	}

	for _, p := range profs {
		t.Log(p)
	}
}
