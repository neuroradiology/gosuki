package mozilla

import (
	"gomark/database"
)

var (
	// user mutable config
	Config *FirefoxConfig

	// Bookmark directory (including profile path)
	bookmarkDir string
)

// Config modifiable by user
type FirefoxConfig struct {
	// Default data source name query options for `places.sqlite` db
	PlacesDSN        database.DsnOptions
	WatchAllProfiles bool
	DefaultProfile   string
}

func SetBookmarkDir(dir string) {
	bookmarkDir = dir
}

func GetBookmarkDir() string {
	return bookmarkDir
}

func SetConfig(c *FirefoxConfig) {
	Config = c
}

func init() {
	Config = new(FirefoxConfig)
}
