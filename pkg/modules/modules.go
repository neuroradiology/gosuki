// Modules allow the extension of gosuki to handle other types of browsers or
// source of data that can be turned into bookmarks.
// 
// # Module Types
//
// 	1. Browsers MUST implement the [BrowserModule] interface.
// 	2. Simple modules MUST implement the [Module] interface.
package modules

import (
	"errors"

	"github.com/urfave/cli/v2"
)

var (
	registeredModules []Module
)

type Context struct {
	Cli *cli.Context
}

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
	ID ModID // Id of this module

	// New returns a pointer to a new instance of a gosuki module.
	// Browser modules MUST implement this method.
	New func() Module
}

type ModID string



func verifyModule(module Module) error {
	var err error

	mod := module.ModInfo()
	if mod.ID == "" {
		err = errors.New("gosuki module ID is missing")
	}
	if mod.New == nil {
		err = errors.New("missing ModInfo.New")
	}
	if val := mod.New(); val == nil {
		err = errors.New("ModInfo.New must return a non-nil module instance")
	}

	return err
}

func RegisterModule(module Module) {
	// do not register browser modules here
	_, bMod := module.(BrowserModule)
	if bMod {
		panic("use RegisterBrowser for browser modules")
	}
	
	if err := verifyModule(module); err != nil {
		panic(err)
	}
	//TODO: Register by ID
	registeredModules = append(registeredModules, module)
}


// Returns a list of registerd browser modules
func GetModules() []Module {
	return registeredModules
}

