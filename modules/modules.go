// Modules allow gosuki to be extended in the future.
// This file should live on it's own package or on the core pacakge
// The goal is to allow a generic interface Module that would allow anything to
// register as a Gosuki module.
//
// Browsers would need to register as gosuki Module and as Browser interfaces
package modules

import (
	"errors"

	"github.com/urfave/cli/v2"
)

var (
	registeredBrowsers []BrowserModule
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

func RegisterBrowser(browserMod BrowserModule) {
	if err := verifyModule(browserMod); err != nil {
		panic(err)
	}

	registeredBrowsers = append(registeredBrowsers, browserMod)
	
	// A browser module is also a module
	registeredModules = append(registeredModules, browserMod)
}

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

func GetBrowserModules() []BrowserModule {
	return registeredBrowsers
}
