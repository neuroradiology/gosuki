package mozilla

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	TestPrefBool   = "test.pref.bool"
	TestPrefNumber = "test.pref.number"
	TestPrefString = "test.pref.string"
	TempFileName   = "prefs-test.js"
)

var (
	TestPrefs = map[string]Pref{
		"BOOL": Pref{
			name:   "test.pref.bool",
			value:  true,
			rawval: "true",
		},
		"TRUE": Pref{
			name:   "test.pref.bool.true",
			value:  true,
			rawval: "true",
		},
		"FALSE": Pref{
			name:   "test.pref.bool.false",
			value:  false,
			rawval: "false",
		},
		"NUMBER": Pref{
			name:   "test.pref.number",
			value:  42,
			rawval: "42",
		},
		"STRING": Pref{
			name:   "test.pref.string",
			value:  "test string",
			rawval: "test string",
		},
	}

	TestPrefsBool = map[string]Pref{
		"TRUE":  TestPrefs["TRUE"],
		"FALSE": TestPrefs["FALE"],
	}

	prefsTempFile *os.File
)

type Pref struct {
	name   string
	value  interface{}
	rawval string
}

func tempFile(name string) *os.File {
	f, err := ioutil.TempFile("", name)

	if err != nil {
		panic(err)
	}

	return f
}

func writeTestPrefFile(f *os.File, p Pref) {
	switch v := p.value.(type) {
	case string:
		fmt.Fprintf(f, "user_pref(\"%s\", \"%s\");\n", p.name, v)
	case bool:
		fmt.Fprintf(f, "user_pref(\"%s\", %t);\n", p.name, v)
	case int:
		fmt.Fprintf(f, "user_pref(\"%s\", %d);\n", p.name, v)
	default:
		fmt.Fprintf(f, "user_pref(\"%s\", %v);\n", p.name, v)

	}

	err := f.Sync()
	if err != nil {
		panic(err)
	}
}

func resetTestPrefFile(f *os.File) {
	err := f.Truncate(0)
	if err != nil {
		panic(err)
	}
}

func TestFindPref(t *testing.T) {
	resetTestPrefFile(prefsTempFile)

	for _, c := range TestPrefs {
		// Write the pref to pref file
		writeTestPrefFile(prefsTempFile, c)

		t.Run(c.name, func(t *testing.T) {
			res, err := FindPref(prefsTempFile.Name(), c.name)
			if err != nil {
				t.Error(err)
			}

			if res != c.rawval {
				t.Fail()
			}
		})
	}
}

func TestGetPrefBool(t *testing.T) {
	resetTestPrefFile(prefsTempFile)

	for _, c := range []string{"TRUE", "FALSE"} {
		writeTestPrefFile(prefsTempFile, TestPrefs[c])

		t.Run(c, func(t *testing.T) {
			res, err := GetPrefBool(prefsTempFile.Name(), TestPrefs[c].name)
			if err != nil {
				t.Error(err)
			}

			if res != TestPrefs[c].value {
				t.Fail()
			}
		})
	}

	// Not a boolean
	writeTestPrefFile(prefsTempFile, TestPrefs["STRING"])
	t.Run("NOTBOOL", func(t *testing.T) {

		_, err := GetPrefBool(prefsTempFile.Name(), TestPrefs["STRING"].name)
		if err != nil &&
			err.Error() != "not a bool" {
			t.Error(err)
		}

	})
}

func TestSetPrefBool(t *testing.T) {
	resetTestPrefFile(prefsTempFile)

	// Write some data to test the append behavior
	writeTestPrefFile(prefsTempFile, TestPrefs["STRING"])

	setVal, _ := TestPrefs["TRUE"].value.(bool)

	err := SetPrefBool(prefsTempFile.Name(), TestPrefs["TRUE"].name, setVal)

	if err != nil {
		t.Error(err)
	}

	res, err := GetPrefBool(prefsTempFile.Name(), TestPrefs["TRUE"].name)
	if err != nil {
		t.Error(err)
	}

	if res != setVal {
		t.Fail()
	}
}

func TestMain(m *testing.M) {

	prefsTempFile = tempFile(TempFileName)

	code := m.Run()

	prefsTempFile.Close()
	os.Remove(prefsTempFile.Name())

	os.Exit(code)
}
