package browsers

import (
	"fmt"
	"path"
	"path/filepath"

	"git.sp4ke.xyz/sp4ke/gomark/database"
	"git.sp4ke.xyz/sp4ke/gomark/logging"
	"git.sp4ke.xyz/sp4ke/gomark/parsing"
	"git.sp4ke.xyz/sp4ke/gomark/tree"
	"git.sp4ke.xyz/sp4ke/gomark/utils"
	"git.sp4ke.xyz/sp4ke/gomark/watch"
	"github.com/sp4ke/hashmap"
)

type BrowserType uint8

// Browser types
const (
	// Chromium based browsers (chrome, brave ... )
	TChrome BrowserType = iota

	// Firefox based browsers ie. they relay on places.sqlite
	TFirefox

	// Other
	TCustom
)

// reducer channel length, bigger means less sensitivity to events
var (
    log = logging.GetLogger("BASE")
    ReducerChanLen = 1000
)

type Browser interface {
	// Returns a pointer to an initialized browser config
	Config() *BrowserConfig
}

type BrowserConfig struct {
	Name string
	Type BrowserType

	// Directory where the bookmark file is stored
	BkDir string

	// Name of bookmarks file
	BkFile string

	WatchedPaths []string

	// In memory sqlite db (named `memcache`).
	// Used to keep a browser's state of bookmarks across jobs.
	BufferDB *database.DB

	// Fast query db using an RB-Tree hashmap.
	// It represents a URL index of the last running job
	URLIndex *hashmap.RBTree

	// Pointer to the root of the node tree
	// The node tree is built again for every Run job on a browser
	NodeTree *tree.Node
	// Various parsing and timing stats
	*parsing.Stats

	watcher        *watch.WatchDescriptor
	UseFileWatcher bool

	parseHooks []parsing.Hook
}

func (browserconfig *BrowserConfig) Watcher() *watch.WatchDescriptor {
	return browserconfig.watcher
}

func (c BrowserConfig) BookmarkPath() (string, error) {
	bPath, err := filepath.EvalSymlinks(path.Join(c.BkDir, c.BkFile))
	if err != nil {
		log.Error(err)
	}

	exists, err := utils.CheckFileExists(bPath)
	if err != nil {
		log.Error(err)
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("not a bookmark path: %s ", bPath)
	}

	return bPath, nil
}

// Browser who implement this interface need to handle all shuttind down and
// cleanup logic in the defined methods. This is usually called at the end of
// the browser instance lifetime
type Shutdowner interface {
	Shutdown() error
}

// Browser who implement this interface will be able to register custom
// hooks which are called during the main Run() to handle commands and
// messages found in tags and parsed data from browsers
type HookRunner interface {
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

// Browser who implement this interface need to handle all shuttind down and
// closing logic in the defined methods. This is usually called at the end of
// the browser instance lifetime
type CleanUpper interface {
	CleanUp() error
}

// Every browser is setup once, the following methods are called in order of
// their corresponding interfaces are implemented.
// TODO!: integrate with refactoring
// 0- Provision: Sets up and custom configiguration to the browser
// 1- Init : any variable and state initialization
// 2- Load: Does the first loading of data (ex first loading of bookmarks )
func Setup(browser BrowserModule) error {

	// Init browser BufferDB

	// Creates in memory Index (RB-Tree)

	log.Infof("setting up browser <%s>", browser.ModInfo().ID)
	browserId := browser.ModInfo().ID

	// Handle Initializers custom Init from Browser module
	initializer, ok := browser.(Initializer)
	if ok {
		initializer.Init()
	}

	// Handle Loaders custom Init from Browser module
	loader, ok := browser.(Loader)
	if ok {
		err := loader.Load()
		if err != nil {

			// If Browser implements Shutdown procedure
			if b, ok := browser.(CleanUpper); ok {
				err = b.CleanUp()
				if err != nil {
					fmt.Println(fmt.Errorf("cleanup <%s>: %v", browserId, err))
					// continue
				}
			}
			log.Errorf("loading <%s>: %v", browserId, err)
			// continue
		}
	}
	return nil
}

// Setup a watcher service using the provided []Watch elements
// Returns true if a new watcher was created. false if it was previously craeted
// or if the browser does not need a watcher (UseFileWatcher == false).
func SetupWatchers(browserConf *BrowserConfig, watches ...*watch.Watch) (bool, error) {
	var err error

	if !browserConf.UseFileWatcher {
		return false, nil
	}

	browserConf.watcher, err = watch.NewWatcher(browserConf.Name, watches...)
	if err != nil {
		return false, err
	}

	return true, nil
}

func SetupWatchersWithReducer(browserConf *BrowserConfig,
	reducerChanLen int,
	watches ...*watch.Watch) (bool, error) {
	var err error

	if !browserConf.UseFileWatcher {
		return false, nil
	}

	browserConf.watcher, err = watch.NewWatcherWithReducer(browserConf.Name, reducerChanLen, watches...)
	if err != nil {
		return false, err
	}

	return true, nil

}

// Used to store bookmark paths and other
// data related to a particular browser kind
// _TODO: replace in chrome with ProfileManager and remove this ref
// type BrowserPaths struct {
// 	BookmarkFile string
// 	BookmarkDir  string
// }
