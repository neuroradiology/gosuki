package cmd

import (
	"github.com/urfave/cli/v2"
)

var modFlags = map[string][]cli.Flag{}

// Register global flags to pass on to the browser module
func RegGlobalFlag(modId string, flag cli.Flag) {
	if flag == nil {
		log.Panic("registering nil flag")
	}

	log.Debugf("<%s> registering global flag: %s=(%v)", modId, flag)
	if _, ok := modFlags[modId]; !ok {
		modFlags[modId] = []cli.Flag{flag}
	} else {
		modFlags[modId] = append(modFlags[modId], flag)
	}
}

// return registered global flags for module
func GlobalFlags(modId string) []cli.Flag {
	return modFlags[modId]
}
