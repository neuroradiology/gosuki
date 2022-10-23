package browsers

import (
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/watch"
)

// Browser interface that every browser plugin needs to implement
type Browser interface {

	// Browser who implement Watchable will be able to register events and
	// callbacks. This is typically used to watch file change events on bookmark
	// files and call the callback Run() to handle the changes.
	watch.Watchable

	// Browser who implement this interface will be able to register custom
	// hooks which are called during the main Run() to handle commands and
	// messages found in tags and parsed data from browsers
	hookRunner

	// Optional interface Loader
	// Optional interface Initializer

	Closer
}

type BrowserID string

// Browser who implement this interface need to handle all shuttind down and
// closing logic in the defined methods. This is usually called at the end of
// the browser instance lifetime
type Closer interface {
	Shutdown()
}

type hookRunner interface {
	RegisterHooks(...parsing.Hook)
}

// Browser who want to load data in a different way than the usual method
// Watchable.Run() method which is auto run on fired watch events should
// implement this interface.
type Loader interface {

	// Load() will be called right after a browser is initialized
	// Return ok, error
	Load() error
}

// Initialize the browser before any data loading or run callbacks
// If a browser wants to do any preparation and prepare custom state before Loader.Load()
// is called and before any Watchable.Run() or other callbacks are executed.
type Initializer interface {

	// Init() is the first method called after a browser instance is created
	// and registered.
	// Return ok, error
	Init() error
}
