package utils

import (
	"os"
	"path/filepath"

	psutil "github.com/shirou/gopsutil/process"
)

func FileProcessUsers(path string) (map[int32]*psutil.Process, error) {
	fusers := make(map[int32]*psutil.Process)

	processes, err := psutil.Processes()
	if err != nil &&
		err != os.ErrPermission {
		return nil, err
	}

	// Eval symlinks
	relPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}

	//log.Debugf("checking against path: %s", relPath)
	for _, p := range processes {

		files, err := p.OpenFiles()
		_, isPathError := err.(*os.PathError)

		if err != nil && isPathError {
			continue
		}

		// Check if path in files

		for _, f := range files {
			//log.Debug(f)
			if f.Path == relPath {
				fusers[p.Pid] = p
			}
		}

	}

	return fusers, nil
}
