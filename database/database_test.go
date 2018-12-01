package database

import (
	"fmt"
	"os"
	"testing"
)

const (
	TestDB = "testdata/gomarkdb_test.sqlite"
)

func TestInitDB(t *testing.T) {
	testDB := &DB{
		Name:       "test",
		Path:       fmt.Sprintf(MemcacheFmt, "test"),
		Handle:     nil,
		EngineMode: DriverDefault,
	}

	//db.In
	// Try to open locked db
	t.Error("test if database is not locked")
	t.Error("test if db path is not found")
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
