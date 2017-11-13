package main

import "log"

func logPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}
