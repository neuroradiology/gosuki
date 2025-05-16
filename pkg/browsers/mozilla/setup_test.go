package mozilla

import (
	"fmt"
	"os"
	"testing"

	"github.com/blob42/gosuki/internal/database"
)

func TestMain(m *testing.M) {
	var err error

	database.RegisterSqliteHooks()

	if prefsTempFile, err = os.CreateTemp(os.TempDir(),
		fmt.Sprintf("%s*", TempFileName)); err != nil {
		os.Exit(1)
	}
	code := m.Run()

	prefsTempFile.Close()
	os.Remove(prefsTempFile.Name())
	os.Exit(code)
}
