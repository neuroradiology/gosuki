package main

import "log"

func init() {
	log.SetFlags(0)
}

func IsDebuggin() bool {
	return goBookmarkMode == debugCode
}

func debugPrint(format string, values ...interface{}) {
	if IsDebuggin() {
		log.Printf("[debug] "+format, values...)
	}
}
