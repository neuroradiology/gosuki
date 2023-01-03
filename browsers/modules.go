// Modules will allow gomark to be extended in the future.
// This file should live on it's own package or on the core pacakge
// The goal is to allow a generic interface Module that would allow anything to
// register as a Gomark module.
//
// Browsers would need to register as gomark Module and as Browser interfaces
package browsers

var (
	registeredBrowsers []BrowserModule
)

// Every new module needs to register as a Module using this interface
type Module interface {
	ModInfo() ModInfo
}

// browser modules need to implement Browser interface
type BrowserModule interface {
	Browser
	Module
}

// Information related to the browser module
type ModInfo struct {
	ID ModID // Id of this browser

	// New returns a pointer to a new instance of a gomark module.
	// Browser modules MUST implement this method.
	New func() Module
}

type ModID string

func RegisterBrowser(browserMod BrowserModule) {
	mod := browserMod.ModInfo()
	if mod.ID == "" {
		panic("gomark module ID is missing")
	}
	if mod.New == nil {
		panic("missing ModInfo.New")
	}
	if val := mod.New(); val == nil {
		panic("ModInfo.New must return a non-nil module instance")
	}

	//TODO: Register by ID
	registeredBrowsers = append(registeredBrowsers, browserMod)
}

// Returns a list of registerd browser modules
func Modules() []BrowserModule {
	return registeredBrowsers
}
