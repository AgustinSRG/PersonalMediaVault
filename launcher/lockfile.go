// Lockfile check

package main

import "github.com/nightlyone/lockfile"

func CheckVaultLocked(file string) bool {
	if !fileExists(file) {
		return false
	}

	f, err := lockfile.New(file)

	if err != nil {
		return true
	}

	err = f.TryLock()

	if err != nil {
		return true
	}

	f.Unlock()

	return false
}
