// Lockfile

package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/nightlyone/lockfile"
)

var (
	VAULT_LOCKFILE lockfile.Lockfile
)

func TryLockVault(base_path string) bool {
	err := os.MkdirAll(base_path, FOLDER_PERMISSION)

	if err != nil {
		LogError(err)
		return false
	}

	lockfile_path, err := filepath.Abs(path.Join(base_path, "vault.lock"))

	if err != nil {
		LogError(err)
		return false
	}

	f, err := lockfile.New(lockfile_path)

	if err != nil {
		LogError(err)
		return false
	}

	err = f.TryLock()

	if err != nil {
		return false
	}

	return true
}
