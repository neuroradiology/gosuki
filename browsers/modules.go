package browsers

var (
	registeredBrowsers []Module
)

// Every new browser plugin needs to register as a browser module using this
// interface
type Module interface {
	ModInfo() ModInfo
}

type BrowserModule interface {
    Browser
    Module
}

// Information related to the browser module
type ModInfo struct {
	ID ModID // Id of this browser

	// New returns a pointer to a new empty instance of gomark browser module
	New func() Module
}

type ModID string

func RegisterBrowser(browserMod Module) {
	registeredBrowsers = append(registeredBrowsers, browserMod)
}

// Returns a list of registerd browser modules
func Modules() []Module {
	return registeredBrowsers
}
