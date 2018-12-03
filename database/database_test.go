package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

const (
	TestDB = "testdata/gomarkdb_test.sqlite"
)

type LockedSQLXOpener struct {
	handle *sqlx.DB
	err    sqlite3.Error
}

func (o *LockedSQLXOpener) Open(driver string, dsn string) error {
	return o.err

}

func (o *LockedSQLXOpener) Get() *sqlx.DB {
	return nil
}

// We
func TestDBLocked(t *testing.T) {
	lockedOpener := &LockedSQLXOpener{
		handle: nil,
		err:    sqlite3.Error{Code: sqlite3.ErrBusy},
	}

	testDB := &DB{
		Name:       "test",
		Path:       fmt.Sprintf(MemcacheFmt, "test"),
		EngineMode: DriverDefault,
		SQLXOpener: lockedOpener,
	}

	_, err := testDB.Init()
	if err != nil {
		e, _ := err.(DBError).Err.(sqlite3.Error)

		if e.Code == sqlite3.ErrBusy {
			t.Error("should handle locked database")
		} else {
			t.Error(err)
		}
	}

}

func TestGomarkDBCeate(t *testing.T) {
	t.Error("if gomark.db does not exist create it")
}

func TestSyncFromGomarkDB(t *testing.T) {
	t.Error("sync from gomark db")
}

func TestSyncToGomarkDB(t *testing.T) {
	t.Error("sync to gomark db")
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
