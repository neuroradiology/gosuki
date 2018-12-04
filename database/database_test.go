package database

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-sqlite3"
)

const (
	TestDB = "testdata/gomarkdb_test.sqlite"
)

func TestNew(t *testing.T) {

	// Test buffer format
	t.Run("BufferPath", func(t *testing.T) {

		db := New("buffer", "", DBTypeInMemoryDSN)

		if db.Path != "file:buffer?mode=memory&cache=shared" {
			t.Error("invalid buffer path")
		}

	})

	t.Run("MemPath", func(t *testing.T) {

		db := New("cache", "", DBTypeCacheDSN)
		if db.Path != "file:cache?mode=memory&cache=shared" {
			t.Fail()
		}

	})

	t.Run("FilePath", func(t *testing.T) {

		db := New("file_test", "/tmp/test/testdb.sqlite", DBTypeFileDSN)

		if db.Path != "file:/tmp/test/testdb.sqlite" {
			t.Fail()
		}

	})

	t.Run("FileCustomDsn", func(t *testing.T) {
		opts := DsnOptions{
			"foo":  "bar",
			"mode": "rw",
		}

		db := New("file_dsn", "", DBTypeFileDSN, opts)

		if db.Path != "file:file_dsn?foo=bar&mode=rw" {
			t.Fail()
		}
	})

	t.Run("AppendOptions", func(t *testing.T) {
		opts := DsnOptions{
			"foo":  "bar",
			"mode": "rw",
		}

		db := New("append_opts", "", DBTypeInMemoryDSN, opts)

		if db.Path != "file:append_opts?mode=memory&cache=shared&foo=bar&mode=rw" {
			t.Fail()
		}
	})
}

type AlwaysLockedChecker struct {
	err error
}

func (f *AlwaysLockedChecker) Locked() (bool, error) {
	return true, nil
}

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

func TestInitLocked(t *testing.T) {
	lockedOpener := &LockedSQLXOpener{
		handle: nil,
		err:    sqlite3.Error{Code: sqlite3.ErrBusy},
	}

	lockChecker := &AlwaysLockedChecker{}

	testDB := &DB{
		Name:        "test",
		Path:        "file:test",
		EngineMode:  DriverDefault,
		SQLXOpener:  lockedOpener,
		Type:        DBTypeRegularFile,
		LockChecker: lockChecker,
	}

	_, err := testDB.Init()

	if err != nil {
		t.Log(err)

		t.Run("VFSLockChecker", func(t *testing.T) {

			t.Error("TODO")

		})

		t.Run("SQLXLockChecker", func(t *testing.T) {

			e, _ := err.(DBError).Err.(sqlite3.Error)

			if e.Code == sqlite3.ErrBusy {
				t.Error("should handle locked database")
			} else {
				t.Fail()
			}
			t.Error("TODO")

		})

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
