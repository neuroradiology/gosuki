package mozilla

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
)
