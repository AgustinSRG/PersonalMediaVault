// Test for credentials manager

package main

import (
	"crypto/subtle"
	"os"
	"path"
	"testing"
)

func TestCreadentialsManager(t *testing.T) {
	test_path_base := "./temp"

	SetTempFilesPath(test_path_base)

	err := os.MkdirAll(test_path_base, FOLDER_PERMISSION)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	test_file := path.Join(test_path_base, "credentials.json")

	// Initalize credentials

	var cred VaultCredentialsManager

	err = cred.Initialize(test_file)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	// Unlock

	key, err := cred.UnlockVault(VAULT_DEFAULT_USER, VAULT_DEFAULT_PASSWORD)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if key == nil {
		t.Errorf("Invalid key received")
	}

	// Change password

	err = cred.SetCredentials("user", "password", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	// Save

	err = cred.SaveCredentials()

	// Create new manager (simulate restart)

	var cred2 VaultCredentialsManager

	cred2.Initialize(test_file)

	// Check password

	b := cred2.CheckCredentials("user", "password")

	if !b {
		t.Errorf("Invalid password, but the password was valid")
	}

	b = cred2.CheckCredentials("user", "passwodd")

	if b {
		t.Errorf("Valid password, but the password was invalid")
	}

	otherKey, err := cred.UnlockVault("user", "password")

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if otherKey == nil {
		t.Errorf("Invalid key received")
	}

	if subtle.ConstantTimeCompare(otherKey, key) != 1 {
		t.Errorf("Keys are different")
	}

	os.Remove(test_file)
}
