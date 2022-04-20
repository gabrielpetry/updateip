package lockfile

import (
	"errors"
	"os"
)

var lockpath = "/tmp/updateip.lock"

func Lock() error {
	if _, err := os.Stat(lockpath); errors.Is(err, os.ErrNotExist) {
		os.Create(lockpath)
		return nil
	}

	panic("lockfile already exists in: " + lockpath)
}

func Unlock() error {
	err := os.Remove(lockpath)

	if err != nil {
		return err
	}

	return nil
}
