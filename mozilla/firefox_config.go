package mozilla

import (
	"gomark/database"
)

var (
	// mutable config
	Config *FirefoxConfig
)

// Default data source name query options for `places.sqlite` db
var PlacesDSN = database.DsnOptions{
	"_journal_mode": "WAL",
}

type FirefoxConfig struct {

	// Bookmark directory (including profile path)
	bookmarkDir string

	WatchAllProfiles bool
	DefaultProfile   string
}

func SetBookmarkDir(dir string) {
	Config.bookmarkDir = dir
}

func GetBookmarkDir() string {
	return Config.bookmarkDir
}

func SetConfig(c *FirefoxConfig) {
	Config = c
}

func init() {
	Config = new(FirefoxConfig)
}
