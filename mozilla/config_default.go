package mozilla

import (
	"gomark/database"
)

var FirefoxDefaultConfig = &FirefoxConfig{

	// Default data source name query options for `places.sqlite` db
	PlacesDSN: database.DsnOptions{
		"_journal_mode": "WAL",
	},

	// default profile to use
	DefaultProfile: "default",

	WatchAllProfiles: false,
}
