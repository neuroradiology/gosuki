package mozilla

import (
	"gomark/logging"
)

const (
	BookmarkFile = "places.sqlite"
)

var (
	log          = logging.GetLogger("FF")
	ConfigFolder = "$HOME/.mozilla/firefox"
)
