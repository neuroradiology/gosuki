package utils

import (
	"os"

	psutil "github.com/shirou/gopsutil/process"
)

func FileProcessUsers(path string) ([]*psutil.Process, error) {
	var fusers []*psutil.Process

	processes, err := psutil.Processes()
	if err != nil &&
		err != os.ErrPermission {
		return nil, err
	}

	for _, p := range processes {

		files, err := p.OpenFiles()
		errPath, _ := err.(*os.PathError)

		if err != nil &&
			errPath.Err.Error() != os.ErrPermission.Error() {
			log.Error(err)
			return nil, err
		}

		// Check if path in files
		for _, f := range files {
			if f.Path == path {
				fusers = append(fusers, p)
			}
		}

	}

	return fusers, nil
}
