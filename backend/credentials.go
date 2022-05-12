// Credentials manager

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
)

const (
	VAULT_CRED_METHOD_AES_SHA256 string = "aes256/sha256/salt16"

	VAULT_DEFAULT_USER     string = "admin"
	VAULT_DEFAULT_PASSWORD string = "admin"
)

type VaultCredentials struct {
	Method       string `json:"method"`
	User         string `json:"user"`
	Salt         []byte `json:"salt"`
	PasswordHash []byte `json:"pwhash"`
	EncryptedKey []byte `json:"enckey"`
}

type VaultCredentialsManager struct {
	file        string
	credentials VaultCredentials

	locked bool
	key    []byte

	lock *sync.Mutex
}

// Loads the creadentials file or creates a new one if
// there is not one
func (manager *VaultCredentialsManager) Initialize(file string) error {
	manager.file = file
	manager.locked = true
	manager.key = nil

	manager.lock = &sync.Mutex{}

	if _, err := os.Stat(file); err == nil {
		// exists
		b, err := ioutil.ReadFile(file)

		if err != nil {
			return err
		}

		// Decrypt
		err = json.Unmarshal(b, &manager.credentials)

		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		rand.Read(key)

		// Set default credentials
		manager.SetCredentials(VAULT_DEFAULT_USER, VAULT_DEFAULT_PASSWORD, key)
		manager.SaveCredentials()
	} else {
		return err
	}

	return nil
}

func (manager *VaultCredentialsManager) SetCredentials(user string, password string, key []byte) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set method
	manager.credentials.Method = VAULT_CRED_METHOD_AES_SHA256

	// Set user
	manager.credentials.User = user

	// Random salt
	manager.credentials.Salt = make([]byte, 16)
	rand.Read(manager.credentials.Salt)

	// Store ecrypted key
	pwBytes := []byte(password)
	ctBytes := make([]byte, len(pwBytes)+16)
	copy(ctBytes[:len(pwBytes)], pwBytes)
	copy(ctBytes[len(pwBytes):], manager.credentials.Salt)
	pwHash := sha256.Sum256(ctBytes)

	encKey, err := encryptFileContents(key, AES256_FLAT, pwHash[:])

	if err != nil {
		return err
	}

	manager.credentials.EncryptedKey = encKey

	// Store password hash
	pwDoubleHash := sha256.Sum256(pwHash[:])

	manager.credentials.PasswordHash = pwDoubleHash[:]

	return nil
}

func (manager *VaultCredentialsManager) SaveCredentials() error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Get the json data
	jsonData, err := json.Marshal(manager.credentials)

	if err != nil {
		return err
	}

	// Make a temp file
	tFile := GetTemporalFileName("json")

	// Write file
	err = ioutil.WriteFile(tFile, jsonData, 0666)
	if err != nil {
		return err
	}

	// Move to the original path
	err = os.Rename(tFile, manager.file)
	if err != nil {
		return err
	}

	return nil
}

func (manager *VaultCredentialsManager) CheckCredentials(user string, password string) bool {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.credentials.User != user {
		return false
	}

	if manager.credentials.Method == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], manager.credentials.Salt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		return subtle.ConstantTimeCompare(pwDoubleHash[:], manager.credentials.PasswordHash) == 1
	} else {
		return false // Unknown method
	}
}

func (manager *VaultCredentialsManager) LockVault() {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.key = nil
	manager.locked = true
}

func (manager *VaultCredentialsManager) UnlockVault(user string, password string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if !manager.locked {
		return nil // Already unlocked
	}

	if manager.credentials.User != user {
		return errors.New("Unknown user")
	}

	if manager.credentials.Method == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], manager.credentials.Salt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		if subtle.ConstantTimeCompare(pwDoubleHash[:], manager.credentials.PasswordHash) != 1 {
			return errors.New("Invalid credentials")
		}

		// Decrypt key
		key, err := decryptFileContents(manager.credentials.EncryptedKey, pwHash[:])

		if err != nil {
			return err
		}

		manager.key = key
		manager.locked = false
		return nil
	} else {
		return errors.New("Unknown credentials method")
	}
}

// Gets key, returns nil if the vault is locked
func (manager *VaultCredentialsManager) GetKey() []byte {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.locked {
		return nil
	}

	return manager.key
}
