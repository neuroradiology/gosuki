package mozilla

import (
	"gomark/logging"
)

var fflog = logging.GetLogger("FF")

const (
	BookmarkFile = "places.sqlite"
	BookmarkDir  = "/home/spike/.mozilla/firefox/7otsk3vs.test_bookmarks"
)

var (
	// Default data source name query options for `places.sqlite` db
	PlacesDSN = map[string]string{
		"_jouranl_mode": "WAL",
	}
	log = logging.GetLogger("MOZ")
)
