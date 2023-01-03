// It is possible to enable debugging for execution time that happens before
// the -debug cli arg is parsed. This is possible using the GOMARK_DEBUG=X env 
// variable where X is an integer for the debug level
package logging

import (
	"os"

	"strconv"

	glogging "github.com/op/go-logging"
)

var (
	log         = glogging.MustGetLogger("MODE")

    //RELEASE: Change to Release for release mode
	LoggingMode = Debug

)

const EnvGomarkDebug = "GOMARK_DEBUG"

const Test = -1
const (
	Release = iota
	Info
	Debug
)

func SetMode(lvl int) {
	if lvl > Debug || lvl < -1 {
		log.Warningf("using wrong debug level: %v", lvl)
		return
	}
	LoggingMode = lvl
    setLogLevel(lvl)
}

func initRuntimeMode() {

	envDebug := os.Getenv(EnvGomarkDebug)

	if envDebug != "" {
		mode, err := strconv.Atoi(os.Getenv(EnvGomarkDebug))

		if err != nil {
			log.Errorf("wrong debug level: %v\n%v", envDebug, err)
		}

        SetMode(mode)
	} 

    //TODO: disable debug log when testing
}
