package mozilla

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
)

const (
    // Note that user.js will be read every time Firefox starts, so changes made to it will take effect immediately.
    // Also when Firefox is updated, it may update also the prefs.js file, and your
    // modification could be lost. Therefore it is better to use user.js to set such
    // preferences which are not possible to set from the Firefox settings.
    //TODO!: create file if it does not exist
	PrefsFile = "user.js"

	// Parses vales in prefs.js under the form:
	// user_pref("my.pref.option", value);
	REFirefoxPrefs = `user_pref\("(?P<option>%s)",\s+"*(?P<value>.*[^"])"*\)\s*;\s*(\n|$)`
)

var (
	ErrPrefNotFound = errors.New("pref not defined")
	ErrPrefNotBool  = errors.New("pref is not bool")
)

// Finds and returns a prefernce definition.
// Returns empty string ("") if no pref found
func FindPref(path string, name string) (string, error) {
	text, err := os.ReadFile(path)
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

// Returns true if the `name` preference is found in `prefs.js`
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

func GetPrefBool(path string, name string) (bool, error) {
	val, err := FindPref(path, name)

	if err != nil {
		return false, err
	}

	switch val {
	case "":
		return false, ErrPrefNotFound
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, ErrPrefNotBool
	}
}

// Set a preference in the preference file under `path`
func SetPrefBool(path string, name string, val bool) error {
	// Get file mode
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()

	// Pref already defined, replace it
	if v, _ := HasPref(path, name); v {

		f, err := os.OpenFile(path, os.O_RDWR, mode)
		defer f.Sync()
		defer f.Close()

		if err != nil {
			return err
		}

		re := regexp.MustCompile(fmt.Sprintf(REFirefoxPrefs, name))
		template := []byte(fmt.Sprintf("user_pref(\"$option\", %t) ;\n", val))
		text, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			return err
		}

		output := string(re.ReplaceAll(text, template))
		fmt.Fprint(f, output)

	} else {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, mode)
		defer f.Sync()
		defer f.Close()

		if err != nil {
			return err
		}

		// Append pref
		fmt.Fprintf(f, "user_pref(\"%s\", %t);\n", name, val)
	}

	return nil
}
