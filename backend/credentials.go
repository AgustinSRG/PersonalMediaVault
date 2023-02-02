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

// Vault account
type VaultCredentialsAccount struct {
	User         string `json:"user"`   // Username
	PasswordHash []byte `json:"pwhash"` // Password hash
	Method       string `json:"method"` // Password hashing / encryption method
	Salt         []byte `json:"salt"`   // Password salt
	EncryptedKey []byte `json:"enckey"` // Vault key encrypted with password

	Write bool `json:"write"` // Write access
}

// Vault credentials data structure
type VaultCredentials struct {
	User         string `json:"user"`   // Root username
	PasswordHash []byte `json:"pwhash"` // Root password hash
	Method       string `json:"method"` // Root password hashing / encryption method
	Salt         []byte `json:"salt"`   // Root password salt
	EncryptedKey []byte `json:"enckey"` // Vault key encrypted with root password

	VaultFingerprint string `json:"fingerprint"` // Vault finderprint, user to distinguis between vaults

	Accounts []VaultCredentialsAccount `json:"accounts"` // Accounts
}

// Credentials managger
type VaultCredentialsManager struct {
	credentials VaultCredentials // Credentials data

	file string // Storage file path

	lock *sync.Mutex // Mutex
}

// Generates a random fingerprint for the vault
func GenerateFingerprint() string {
	data := make([]byte, 16)
	now := time.Now().UnixMilli()

	binary.BigEndian.PutUint64(data[:8], uint64(now))
	rand.Read(data[8:])

	return hex.EncodeToString(data)
}

// Initializes credentials path
// Ask for username and password
// using the standard input
// base_path - Path
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

		var username string = os.Getenv("PMV_INIT_SET_USER")

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

		var password string = os.Getenv("PMV_INIT_SET_PASSWORD")
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

// Initialization
// Loads the credentials file
// or creates a new one if there is not one
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

		if manager.credentials.Accounts == nil {
			manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		rand.Read(key)

		// Set default credentials
		manager.credentials.VaultFingerprint = GenerateFingerprint()
		manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)
		manager.SetRootCredentials(VAULT_DEFAULT_USER, VAULT_DEFAULT_PASSWORD, key)
		manager.SaveCredentials()
	} else {
		return err
	}

	return nil
}

// Creates new credentials file using provided credentials
// file - Credentials file
// user - Root username
// password - Root password
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
		manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)
		manager.SetRootCredentials(user, password, key)
		manager.SaveCredentials()
	} else {
		return err
	}

	return nil
}

// Sets root credentials
// user - Root username
// password - Root password
// key - Vault encryption key
func (manager *VaultCredentialsManager) SetRootCredentials(user string, password string, key []byte) error {
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

// Sets account credentials (creates a new one if not found)
// user - Root username
// password - Root password
// key - Vault encryption key
// write - Write access
func (manager *VaultCredentialsManager) SetAccountCredentials(user string, password string, key []byte, write bool) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	var accountIndex int = -1

	for i := 0; i < len(manager.credentials.Accounts); i++ {
		if manager.credentials.Accounts[i].User == user {
			accountIndex = i
		}
	}

	if accountIndex == -1 {
		manager.credentials.Accounts = append(manager.credentials.Accounts, VaultCredentialsAccount{
			User: user,
		})
		accountIndex = len(manager.credentials.Accounts) - 1
	}

	// Set method
	manager.credentials.Accounts[accountIndex].Method = VAULT_CRED_METHOD_AES_SHA256

	// Random salt
	manager.credentials.Accounts[accountIndex].Salt = make([]byte, 16)
	rand.Read(manager.credentials.Accounts[accountIndex].Salt)

	// Store ecrypted key
	pwBytes := []byte(password)
	ctBytes := make([]byte, len(pwBytes)+16)
	copy(ctBytes[:len(pwBytes)], pwBytes)
	copy(ctBytes[len(pwBytes):], manager.credentials.Accounts[accountIndex].Salt)
	pwHash := sha256.Sum256(ctBytes)

	encKey, err := encryptFileContents(key, AES256_FLAT, pwHash[:])

	if err != nil {
		return err
	}

	manager.credentials.Accounts[accountIndex].EncryptedKey = encKey

	// Store password hash
	pwDoubleHash := sha256.Sum256(pwHash[:])

	manager.credentials.Accounts[accountIndex].PasswordHash = pwDoubleHash[:]

	// Set write access
	manager.credentials.Accounts[accountIndex].Write = write

	return nil
}

// Removes an account
// user - Username
func (manager *VaultCredentialsManager) RemoveAccount(user string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	newAccounts := make([]VaultCredentialsAccount, 0)

	for i := 0; i < len(manager.credentials.Accounts); i++ {
		if manager.credentials.Accounts[i].User != user {
			newAccounts = append(newAccounts, manager.credentials.Accounts[i])
		}
	}

	manager.credentials.Accounts = newAccounts

	return nil
}

// Changes username
// user - Username
// new_user - New username
func (manager *VaultCredentialsManager) ChangeUsername(user string, new_user string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	if manager.credentials.User == user {
		manager.credentials.User = new_user
	} else {
		for i := 0; i < len(manager.credentials.Accounts); i++ {
			if manager.credentials.Accounts[i].User == user {
				manager.credentials.Accounts[i].User = new_user
				break
			}
		}
	}

	return nil
}

// Saves credentials data to the vault permanent storage
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

// Credentials check result
type CredentialsCheckResult struct {
	root  bool // Root account
	write bool // Write access
}

// Checks credentials
// user - Username
// password - Password
// Returns (success, result)
// - success is true if the credentails chccek succedded
// - result contains info if the credentials are root credentials and if the have write access
func (manager *VaultCredentialsManager) CheckCredentials(user string, password string) (bool, *CredentialsCheckResult) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	result := CredentialsCheckResult{
		root:  false,
		write: false,
	}

	var pwMethod string
	var pwSalt []byte
	var pwExpectedHash []byte

	if manager.credentials.User == user {
		result.root = true
		result.write = true
		pwMethod = manager.credentials.Method
		pwSalt = manager.credentials.Salt
		pwExpectedHash = manager.credentials.PasswordHash
	} else {
		result.root = false
		var foundAccount bool = false
		for i := 0; i < len(manager.credentials.Accounts); i++ {
			if manager.credentials.Accounts[i].User == user {
				foundAccount = true
				result.write = manager.credentials.Accounts[i].Write
				pwMethod = manager.credentials.Accounts[i].Method
				pwSalt = manager.credentials.Accounts[i].Salt
				pwExpectedHash = manager.credentials.Accounts[i].PasswordHash
				break
			}
		}

		if !foundAccount {
			return false, nil
		}
	}

	if pwMethod == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], pwSalt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		checkResult := subtle.ConstantTimeCompare(pwDoubleHash[:], pwExpectedHash) == 1

		if checkResult {
			return true, &result
		} else {
			return false, nil
		}
	} else {
		return false, nil // Unknown method
	}
}

// Unlocks vaults and gets decryption key
// user - Username
// password - Password
// Returns the key and the result (root and write info)
// If fails, it returns an error
func (manager *VaultCredentialsManager) UnlockVault(user string, password string) ([]byte, *CredentialsCheckResult, error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	result := CredentialsCheckResult{
		root:  false,
		write: false,
	}

	var pwMethod string
	var pwSalt []byte
	var pwExpectedHash []byte
	var pwEcryptedKey []byte

	if manager.credentials.User == user {
		result.root = true
		result.write = true
		pwMethod = manager.credentials.Method
		pwSalt = manager.credentials.Salt
		pwExpectedHash = manager.credentials.PasswordHash
		pwEcryptedKey = manager.credentials.EncryptedKey
	} else {
		result.root = false
		var foundAccount bool = false
		for i := 0; i < len(manager.credentials.Accounts); i++ {
			if manager.credentials.Accounts[i].User == user {
				foundAccount = true
				result.write = manager.credentials.Accounts[i].Write
				pwMethod = manager.credentials.Accounts[i].Method
				pwSalt = manager.credentials.Accounts[i].Salt
				pwExpectedHash = manager.credentials.Accounts[i].PasswordHash
				pwEcryptedKey = manager.credentials.Accounts[i].EncryptedKey
				break
			}
		}

		if !foundAccount {
			return nil, nil, errors.New("Unknown user")
		}
	}

	if pwMethod == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], pwSalt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		if subtle.ConstantTimeCompare(pwDoubleHash[:], pwExpectedHash) != 1 {
			return nil, nil, errors.New("Invalid credentials")
		}

		// Decrypt key
		key, err := decryptFileContents(pwEcryptedKey, pwHash[:])

		if err != nil {
			return nil, nil, err
		}

		return key, &result, nil
	} else {
		return nil, nil, errors.New("Unknown credentials method")
	}
}

// Gets vault fingerpint
func (manager *VaultCredentialsManager) GetFingerprint() string {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	return manager.credentials.VaultFingerprint
}

// Checks if user exists
// user - Username
// Returns true if the user exists
func (manager *VaultCredentialsManager) CheckUserExists(user string) bool {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if manager.credentials.User == user {
		return true
	}

	for i := 0; i < len(manager.credentials.Accounts); i++ {
		if manager.credentials.Accounts[i].User == user {
			return true
		}
	}

	return false
}

// Gets information of all accounts
func (manager *VaultCredentialsManager) GetAccountsInfo() []ApiAdminAccountEntry {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	result := make([]ApiAdminAccountEntry, 0)

	for i := 0; i < len(manager.credentials.Accounts); i++ {
		result = append(result, ApiAdminAccountEntry{
			Username: manager.credentials.Accounts[i].User,
			Write:    manager.credentials.Accounts[i].Write,
		})
	}

	return result
}
