package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"git.blob42.xyz/gomark/gosuki/logging"
)

var (
	TMPDIR = ""
	log    = logging.GetLogger("")
)

func copyFileToDst(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil

}

// Copy files from src glob to dst folder
func CopyFilesToTmpFolder(srcglob string, dst string) error {
	matches, err := filepath.Glob(os.ExpandEnv(srcglob))
	if err != nil {
		return err
	}

	for _, v := range matches {
		dstFile := path.Join(dst, path.Base(v))
		err = copyFileToDst(v, dstFile)
		if err != nil {
			return err
		}

	}

	return nil

}

func CleanFiles() {
	log.Debugf("Cleaning files <%s>", TMPDIR)
	err := os.RemoveAll(TMPDIR)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	var err error
	TMPDIR, err = ioutil.TempDir("", "gosuki*")
	if err != nil {
		log.Fatal(err)
	}
}
