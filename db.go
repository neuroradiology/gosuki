package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db *sql.DB
)

const (
	CREATE_DB_SCHEMA = `CREATE TABLE if not exists bookmarks (
		id integer PRIMARY KEY,
		URL text NOT NULL UNIQUE)`
)

// TODO: Use context when making call from request/api

func initInMemoryDb() {

	var err error
	//var tx *sql.Tx

	db, err = sql.Open("sqlite3", "file::memory:?cache=shared")
	debugPrint("in memory db opened")
	logPanic(err)

	_, err = db.Exec(CREATE_DB_SCHEMA)
	logPanic(err)

	//tx, err = db.Begin()
	//logPanic(err)

	//tx.Exec(CREATE_DB_SCHEMA)

	//tx.Commit()
}

func initdb() {
	debugPrint("[NotImplemented] initialize local db if not exists or load it")
	// Check if db exists locally

	// If does not exit create new one in memory
	// Then flush to disk

	// If exists locally, load to memory
	initInMemoryDb()
}

func testInMemoryDb() {

	debugPrint("test in memory")
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	defer db.Close()
	rows, err := db.Query("select URL from bookmarks")
	defer rows.Close()
	logPanic(err)
	var URL string
	for rows.Next() {
		rows.Scan(&URL)
		log.Println(URL)
	}
}
