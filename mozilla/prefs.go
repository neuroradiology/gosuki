package mozilla

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

const (
	PrefsFile = "prefs.js"

	// Parses vales in prefs.js under the form:
	// user_pref("my.pref.option", value);
	REFirefoxPrefs = `user_pref\("(?P<option>%s)",\s+"*(?P<value>.*[^"])"*\);`
)

// Finds and returns a prefernce definition.
// Returns empty string ("") if no pref found
func FindPref(path string, name string) (string, error) {
	text, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(fmt.Sprintf(REFirefoxPrefs, name))
	match := re.FindSubmatch(text)

	if match == nil {
		return "", nil
	}

	results := map[string]string{}
	for i, name := range match {
		results[re.SubexpNames()[i]] = string(name)
	}

	return results["value"], nil
}

func HasPref(path string, name string) (bool, error) {
	res, err := FindPref(path, name)
	if err != nil {
		return false, err
	}

	if res != "" {
		return true, nil
	}

	return false, nil
}

//TODO:  Pass profile name and use standard `prefs.js` file name in base
// directory of profile
func GetPrefBool(path string, name string) (bool, error) {
	val, err := FindPref(path, name)

	if err != nil {
		return false, err
	}

	if val == "" {
		return false, errors.New("not found")
	}

	if val == "true" {
		return true, nil
	} else if val == "false" {
		return false, nil
	}

	return false, errors.New("not a bool")

}

// Set a preference in the preference file under `path`
func SetPrefBool(path string, name string, val bool) error {
	// Get file mode
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()

	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, mode)
	defer f.Close()

	if err != nil {
		return err
	}

	fmt.Println(name, val)
	fmt.Fprintf(f, "user_pref(\"%s\", %t);\n", name, val)
	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}
