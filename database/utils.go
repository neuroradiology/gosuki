package database

import (
    _ "io"
    "embed"
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

// Loads a dotsql from an embedded FS
func DotxQueryEmbedFS(fs embed.FS, filename string) (*dotsqlx.DotSqlx, error){

    rawsql, err := fs.ReadFile(filename)
    if err != nil {
      return nil, err
    }
    

    dot, err := dotsql.LoadFromString(string(rawsql))
    if err != nil {
      return nil, err
    }

    return dotsqlx.Wrap(dot), nil
}
