package browsers

var (
	registeredBrowsers []BrowserMod
)

// Every new browser plugin needs to register it's ID using this struct
type BrowserMod struct {
	Name    string
	Browser Browser
}

func Register(browser BrowserMod) {
	registeredBrowsers = append(registeredBrowsers, browser)
}

// Returns a list of registerd browser modules
func Modules() []BrowserMod {
	return registeredBrowsers
}
