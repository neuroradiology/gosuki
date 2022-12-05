package database

import (
	"path/filepath"

	"github.com/gchaincl/dotsql"
	"github.com/swithek/dotsqlx"
)

func GetDefaultDBPath() string {
	return DefaultDBPath
}

func GetDBFullPath() string {
	dbdir := GetDefaultDBPath()
	dbpath := filepath.Join(dbdir, DBFileName)
	return dbpath
}

// Loads a dotsql <file> and, wraps it with dotsqlx 
func DotxQuery(file string) (*dotsqlx.DotSqlx, error){
    dot, err := dotsql.LoadFromFile(file)
    if err != nil {
      return nil, err
    }

    return dotsqlx.Wrap(dot), nil
}
