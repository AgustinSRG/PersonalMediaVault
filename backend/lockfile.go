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

// Locks the vault
// Prevents multiple processes to lock the same vault
// base_path - Vault path
// Returns true if vault was locked, false if already locked by other process
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

	return err == nil
}
