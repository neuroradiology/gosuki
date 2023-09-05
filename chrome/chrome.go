package chrome

import (
	"git.blob42.xyz/gomark/gosuki/modules"
)

// Chrome browser module
type Chrome struct {

	// holds the browsers.BrowserConfig
	*ChromeConfig
}

// Returns a pointer to an initialized browser config
func (c Chrome) Config() *modules.BrowserConfig {
	return c.BrowserConfig
}

func (c Chrome) ModInfo() modules.ModInfo {
	return modules.ModInfo{
		ID: modules.ModID(c.Name),
		New: func() modules.Module {
			return NewChrome()
		},
	}
}

func NewChrome() *Chrome {
	return &Chrome{
		ChromeConfig: ChromeCfg,
	}
}

func init(){
  modules.RegisterBrowser(Chrome{ChromeConfig: ChromeCfg})
}



