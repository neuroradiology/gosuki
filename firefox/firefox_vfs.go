// TODO: vfs for firefox should be removed in favor of copy/parse method
package firefox

import (
	"errors"
	"fmt"
	"path"

	moz "git.sp4ke.xyz/sp4ke/gomark/mozilla"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
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
	ErrMultiProcessAlreadyEnabled = errors.New("multiProcessAccess already enabled")
)

func CheckVFSLock() error {
	log.Debugf("Checking VFS for <%s>", getBookmarkDir())
	return nil
}

func UnlockPlaces() error {
	log.Debugf("Unlocking VFS <%s>", path.Join(getBookmarkDir(), moz.PrefsFile))

	prefsPath := path.Join(getBookmarkDir(), moz.PrefsFile)

	// Find if multiProcessAccess option is defined

	pref, err := moz.GetPrefBool(prefsPath, PrefMultiProcessAccess)
	if err != nil && err != moz.ErrPrefNotFound {
		return err
	}

	// If pref already defined and true raise an error
	if pref {
		log.Warningf("pref <%s> already defined as <%v>",
			PrefMultiProcessAccess, pref)
		return nil

		// Set the preference
	} else {

		// Checking if firefox is running
		// TODO: #multiprocess add CLI to unlock places.sqlite
		pusers, err := utils.FileProcessUsers(path.Join(getBookmarkDir(), moz.BookmarkFile))
		if err != nil {
			log.Error(err)
		}

		for pid, p := range pusers {
			pname, err := p.Name()
			if err != nil {
				log.Error(err)
			}
			return errors.New(fmt.Sprintf("multiprocess not enabled and %s(%d) is running. Close firefox and disable VFS lock", pname, pid))
		}
		// End testing

		// enable multi process access in firefox
		err = moz.SetPrefBool(prefsPath,
			PrefMultiProcessAccess,
			true)

		if err != nil {
			return err
		}

	}

	return nil

}
