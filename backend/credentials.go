// Credentials manager

package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	VAULT_CRED_METHOD_AES_SHA256 string = "aes256/sha256/salt16"

	VAULT_DEFAULT_USER     string = "admin"
	VAULT_DEFAULT_PASSWORD string = "admin"
)

type VaultCredentials struct {
	Method           string `json:"method"`
	User             string `json:"user"`
	Salt             []byte `json:"salt"`
	PasswordHash     []byte `json:"pwhash"`
	EncryptedKey     []byte `json:"enckey"`
	VaultFingerprint string `json:"fingerprint"`
}

type VaultCredentialsManager struct {
	file        string
	credentials VaultCredentials

	lock *sync.Mutex
}

func GenerateFingerprint() string {
	data := make([]byte, 16)
	now := time.Now().UnixMilli()

	binary.BigEndian.PutUint64(data[:8], uint64(now))
	rand.Read(data[8:])

	return hex.EncodeToString(data)
}

func InitializeCredentialsPath(base_path string) {
	err := os.MkdirAll(base_path, FOLDER_PERMISSION)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	file := path.Join(base_path, "credentials.json")

	if _, err := os.Stat(file); err == nil {
		fmt.Println("Vault already exists. Skipping initializing process.")
		return
	} else if errors.Is(err, os.ErrNotExist) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Vault does not exists. Please provide a set of credentials to create one.")

		var username string = ""

		for username == "" {
			fmt.Print("Enter Username: ")
			username, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}

			username = strings.TrimSpace(username)

			if username == "" {
				fmt.Println("Username cannot be blank.")
				continue
			}

			if len(username) > 255 {
				fmt.Println("Username cannot be longer than 255 characters.")
				username = ""
				continue
			}
		}

		var password string
		var password_repeat string

		for password == "" || password != password_repeat {
			fmt.Print("Enter Password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}

			password = strings.TrimSpace(string(bytePassword))

			if password == "" {
				fmt.Println("Password cannot be blank.")
				continue
			}

			if len(password) > 255 {
				fmt.Println("Password cannot be longer than 255 characters.")
				password = ""
				continue
			}

			fmt.Print("\n")

			fmt.Print("Repeat Password: ")
			bytePassword, err = term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}

			fmt.Print("\n")

			password_repeat = strings.TrimSpace(string(bytePassword))

			if password != password_repeat {
				fmt.Println("Passwords do not match.")
			}
		}

		cm := VaultCredentialsManager{}

		err = cm.Create(file, username, password)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		fmt.Println("Vault initialized successfully!")
	} else {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}
}

// Loads the creadentials file or creates a new one if
// there is not one
func (manager *VaultCredentialsManager) Initialize(base_path string) error {
	manager.file = path.Join(base_path, "credentials.json")

	manager.lock = &sync.Mutex{}

	if _, err := os.Stat(manager.file); err == nil {
		// exists
		b, err := ioutil.ReadFile(manager.file)

		if err != nil {
			return err
		}

		// Parse
		err = json.Unmarshal(b, &manager.credentials)

		if err != nil {
			return err
		}

		if manager.credentials.VaultFingerprint == "" {
			manager.credentials.VaultFingerprint = GenerateFingerprint()
			manager.SaveCredentials()
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		rand.Read(key)

		// Set default credentials
		manager.credentials.VaultFingerprint = GenerateFingerprint()
		manager.SetCredentials(VAULT_DEFAULT_USER, VAULT_DEFAULT_PASSWORD, key)
		manager.SaveCredentials()
	} else {
		return err
	}

	return nil
}

// Creates new credentials file using provided credentials
func (manager *VaultCredentialsManager) Create(file string, user string, password string) error {
	manager.file = file

	manager.lock = &sync.Mutex{}

	if _, err := os.Stat(file); err == nil {
		// exists
		return errors.New("There is already an existing vault in the provided path.")
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		rand.Read(key)

		// Set default credentials
		manager.credentials.VaultFingerprint = GenerateFingerprint()
		manager.SetCredentials(user, password, key)
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

func (manager *VaultCredentialsManager) ChangeUsername(user string, new_user string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	manager.credentials.User = new_user

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
	tFile := GetTemporalFileName("json", true)

	// Write file
	err = ioutil.WriteFile(tFile, jsonData, FILE_PERMISSION)
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

// Unlocks vaults and gets decryption key
func (manager *VaultCredentialsManager) UnlockVault(user string, password string) ([]byte, error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.credentials.User != user {
		return nil, errors.New("Unknown user")
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
			return nil, errors.New("Invalid credentials")
		}

		// Decrypt key
		key, err := decryptFileContents(manager.credentials.EncryptedKey, pwHash[:])

		if err != nil {
			return nil, err
		}

		return key, nil
	} else {
		return nil, errors.New("Unknown credentials method")
	}
}

func (manager *VaultCredentialsManager) GetFingerprint() string {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	return manager.credentials.VaultFingerprint
}
