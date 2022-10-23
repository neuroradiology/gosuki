package database

import (
	"path/filepath"
)

func GetDefaultDBPath() string {
	return DefaultDBPath
}

func GetDBFullPath() string {
	dbdir := GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
	return dbpath
}
