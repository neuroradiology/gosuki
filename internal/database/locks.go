package database

import (
	"os"

	"golang.org/x/sys/unix"
)

type LockChecker interface {
	Locked() (bool, error)
}

type VFSLockChecker struct {
	path string
}

func (checker *VFSLockChecker) Locked() (bool, error) {

	f, err := os.Open(checker.path)
	if err != nil {
		return false, err
	}

	// Get the the lock mode
	var lock unix.Flock_t
	// See man (fcntl)
	unix.FcntlFlock(f.Fd(), unix.F_GETLK, &lock)

	// Check if lock is F_RDLCK (non-exclusive) or F_WRLCK (exclusive)
	if lock.Type == unix.F_RDLCK {
		//fmt.Println("Lock is F_RDLCK")
		return false, nil
	}

	if lock.Type == unix.F_WRLCK {
		//fmt.Println("Lock is F_WRLCK (locked !)")
		return true, nil
	}

	return false, nil

}
