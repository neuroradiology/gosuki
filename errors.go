package main

import (
	"log"
)

func logPanic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func logError(err error, msg string) {
	if err != nil {
		log.Printf("%s : %s", err, msg)
	}
}
