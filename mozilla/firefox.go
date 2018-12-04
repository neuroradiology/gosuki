package mozilla

import (
	"errors"
	"gomark/logging"
	"path"
)

const (
	// This option disables the VFS lock on firefox
	// Sqlite allows file locking of the database using the local file system VFS.
	// Previous versions of FF allowed external processes to access the file.
	//
	// Since firefox v(63) this has changed. When initializing the database FF checks
	// the preference option `storage.multiProcessAccess.enabled` which is not
	// documented officially.
	//
	// Source code:
	//- https://dxr.mozilla.org/mozilla-central/source/storage/TelemetryVFS.cpp#884
	//- https://dxr.mozilla.org/mozilla-central/source/storage/mozStorageService.cpp#377
	//- Change on github: https://github.com/mozilla/gecko-dev/commit/a543f35d4be483b19446304f52e4781d7a4a0a2f
	PrefMultiProcessAccess = "storage.multiProcessAccess.enabled"
)

var (
	// Default data source name query options for `places.sqlite` db
	PlacesDSN = map[string]string{
		"_jouranl_mode": "WAL",
	}
	log = logging.GetLogger("MOZ")
)

var (
	ErrMultiProcessAlreadyEnabled = errors.New("multiProcessAccess already enabled")
)

//TODO: try unlock at the browser level !
// Try to unlock vfs locked places.sqlite by setting the `PrefMultiProcessAccess`
// property in prefs.js

func UnlockPlaces(dir string) error {
	log.Debug("Unlocking ...")

	prefsPath := path.Join(dir, PrefsFile)

	// Find if multiProcessAccess option is defined

	pref, err := GetPrefBool(prefsPath, PrefMultiProcessAccess)
	if err != nil && err != ErrPrefNotFound {
		return err
	}

	// If pref already defined and true raise an error
	if pref {
		log.Criticalf("pref <%s> already defined as <%s>",
			PrefMultiProcessAccess, pref)
		return ErrMultiProcessAlreadyEnabled

		// Set the preference
	} else {
		log.Debug("pref not defined")

		// enable multi process access in firefox
		err = SetPrefBool(prefsPath,
			PrefMultiProcessAccess,
			true)

		if err != nil {
			return err
		}

	}

	return nil

}
