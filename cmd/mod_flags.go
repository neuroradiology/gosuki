package cmd

import (
	"github.com/urfave/cli/v2"
)

var modFlags = map[string][]cli.Flag{}

// RegGlobalModFlag registers global flags to pass on to the browser module
func RegGlobalModFlag(modID string, flag cli.Flag) {
	if flag == nil {
		log.Panic("registering nil flag")
	}

	log.Debugf("<%s> registering global flag: %s",
				modID,
				flag.Names())
	if _, ok := modFlags[modID]; !ok {
		modFlags[modID] = []cli.Flag{flag}
	} else {
		modFlags[modID] = append(modFlags[modID], flag)
	}
}

// GlobalFlags returns the registered global flags for a given registered module
func GlobalFlags(modID string) []cli.Flag {
	return modFlags[modID]
}
