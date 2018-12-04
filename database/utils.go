package database

import (
	"gomark/utils"
	"path/filepath"
)

func GetDBFullPath() string {
	dbdir := utils.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
	return dbpath
}
