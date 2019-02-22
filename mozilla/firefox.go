package mozilla

import (
	"gomark/logging"
)

const (
	BookmarkFile = "places.sqlite"
)

var (
	// Default data source name query options for `places.sqlite` db
	PlacesDSN = map[string]string{
		"_jouranl_mode": "WAL",
	}

	log = logging.GetLogger("FF")

	// Bookmark directory
	BookmarkDir string
)

func SetBookmarkDir(dir string) {
	log.Debugf("setting bookmarks dir to <%s>", dir)
	BookmarkDir = dir
}
