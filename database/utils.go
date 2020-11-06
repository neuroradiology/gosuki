package database

import (
	"path/filepath"

	"git.sp4ke.xyz/sp4ke/gomark/utils"
)

func GetDBFullPath() string {
	dbdir := utils.GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
	return dbpath
}
