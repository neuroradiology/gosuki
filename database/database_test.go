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
	locked bool
}

func (f *AlwaysLockedChecker) Locked() (bool, error) {
	return f.locked, nil
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

	lockCheckerTrue := &AlwaysLockedChecker{locked: true}
	lockCheckerFalse := &AlwaysLockedChecker{locked: false}

	t.Run("VFSLockChecker", func(t *testing.T) {

		testDB := &DB{
			Name:        "test",
			Path:        "file:test",
			EngineMode:  DriverDefault,
			LockChecker: lockCheckerTrue,
			SQLXOpener:  lockedOpener,
			Type:        DBTypeRegularFile,
		}

		_, err := testDB.Init()

		if err == nil {
			t.Fail()
		}

		if err != DBErr(testDB.Name, ErrVfsLocked) {
			t.Fail()
		}

	})

	t.Run("SQLXLockChecker", func(t *testing.T) {

		testDB := &DB{
			Name:        "test",
			Path:        "file:test",
			EngineMode:  DriverDefault,
			LockChecker: lockCheckerFalse,
			SQLXOpener:  lockedOpener,
			Type:        DBTypeRegularFile,
		}

		_, err := testDB.Init()

		if err == nil {
			t.Fail()
		}

		e, _ := err.(DBError).Err.(sqlite3.Error)

		if e.Code != sqlite3.ErrBusy {
			t.Fail()
		}

	})

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
