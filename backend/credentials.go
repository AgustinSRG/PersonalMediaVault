// Credentials manager

// cSpell:ignore pwhash, enckey

package main

import (
	"bufio"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
	"golang.org/x/term"
)

const (
	VAULT_DEFAULT_USER     string = "admin"
	VAULT_DEFAULT_PASSWORD string = "admin"

	DEFAULT_AUTH_CONFIRMATION_PERIOD = 120
)

// Vault account
type VaultCredentialsAccount struct {
	User string `json:"user"` // Username

	PasswordHash []byte `json:"pwhash"` // Password hash
	Method       string `json:"method"` // Password hashing / encryption method
	Salt         []byte `json:"salt"`   // Password salt

	EncryptedKey []byte `json:"enckey"` // Vault key encrypted with password

	TwoFactorAuthEnabled      bool   `json:"tfa,omitempty"`        // Two factor auth enabled
	TwoFactorAuthMethod       string `json:"tfa_method,omitempty"` // Two factor auth method
	TwoFactorAuthEncryptedKey []byte `json:"tfa_enckey,omitempty"` // Encrypted TFA key

	AuthConfirmationEnabled *bool   `json:"auth_confirmation,omitempty"`        // Enable auth confirmation?
	AuthConfirmationMethod  string  `json:"auth_confirmation_method,omitempty"` // Method for auth confirmation?
	AuthConfirmationPeriod  *uint32 `json:"auth_confirmation_period,omitempty"` // Auth confirmation period, in seconds. Default: 120

	Write bool `json:"write"` // Write access
}

// Vault credentials data structure
type VaultCredentials struct {
	User string `json:"user"` // Root username

	PasswordHash []byte `json:"pwhash"` // Root password hash
	Method       string `json:"method"` // Root password hashing / encryption method
	Salt         []byte `json:"salt"`   // Root password salt

	EncryptedKey []byte `json:"enckey"` // Vault key encrypted with root password

	TwoFactorAuthEnabled      bool   `json:"tfa,omitempty"`        // Two factor auth enabled
	TwoFactorAuthMethod       string `json:"tfa_method,omitempty"` // Two factor auth method
	TwoFactorAuthEncryptedKey []byte `json:"tfa_enckey,omitempty"` // Encrypted TFA key

	AuthConfirmationEnabled *bool   `json:"auth_confirmation,omitempty"`        // Enable auth confirmation?
	AuthConfirmationMethod  string  `json:"auth_confirmation_method,omitempty"` // Method for auth confirmation?
	AuthConfirmationPeriod  *uint32 `json:"auth_confirmation_period,omitempty"` // Auth confirmation period, in seconds. Default: 120

	VaultFingerprint string `json:"fingerprint"` // Vault fingerprint, user to distinguish between vaults

	Accounts []VaultCredentialsAccount `json:"accounts"` // Accounts
}

// Credentials manager
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
	rand.Read(data[8:]) //nolint:errcheck

	return hex.EncodeToString(data)
}

// Gets confirmation period as duration
func GetAuthConfirmationPeriod(p *uint32) time.Duration {
	if p != nil {
		return time.Duration(*p) * time.Second
	} else {
		return DEFAULT_AUTH_CONFIRMATION_PERIOD * time.Second
	}
}

// Gets confirmation enabled setting
func GetAuthConfirmationEnabled(e *bool) bool {
	if e != nil {
		return *e
	} else {
		return true
	}
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
			readUsername, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}

			username = strings.TrimSpace(readUsername)

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
		var password_repeat string = os.Getenv("PMV_INIT_SET_PASSWORD")

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
		b, err := os.ReadFile(manager.file)

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
			err = manager.SaveCredentials()

			if err != nil {
				return err
			}
		}

		if manager.credentials.Accounts == nil {
			manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		_, err2 := rand.Read(key)

		if err2 != nil {
			return err2
		}

		// Set default credentials
		manager.credentials.VaultFingerprint = GenerateFingerprint()
		manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)

		initialUser := os.Getenv("VAULT_INITIAL_USER")

		if initialUser == "" {
			initialUser = VAULT_DEFAULT_USER
		}

		initialPassword := os.Getenv("VAULT_INITIAL_PASSWORD")

		if initialPassword == "" {
			initialPassword = VAULT_DEFAULT_PASSWORD
		}

		err2 = manager.InitRootCredentials(initialUser, initialPassword, key)

		if err2 != nil {
			return err2
		}

		err2 = manager.SaveCredentials()

		if err2 != nil {
			return err2
		}
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
		return errors.New("there is already an existing vault in the provided path")
	} else if errors.Is(err, os.ErrNotExist) {
		// does *not* exist

		// Create a random key
		key := make([]byte, 32)
		_, err2 := rand.Read(key)

		if err2 != nil {
			return err2
		}

		// Set default credentials
		manager.credentials.VaultFingerprint = GenerateFingerprint()
		manager.credentials.Accounts = make([]VaultCredentialsAccount, 0)
		err2 = manager.InitRootCredentials(user, password, key)

		if err2 != nil {
			return err2
		}

		err2 = manager.SaveCredentials()

		if err2 != nil {
			return err2
		}
	} else {
		return err
	}

	return nil
}

// Initializes root credentials
// user - Root username
// password - Root password
// key - Vault encryption key
func (manager *VaultCredentialsManager) InitRootCredentials(user string, password string, key []byte) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set method
	manager.credentials.Method = VAULT_CRED_METHOD_AES_SHA256

	// Set user
	manager.credentials.User = user

	// Random salt
	manager.credentials.Salt = make([]byte, 16)
	_, err := rand.Read(manager.credentials.Salt)

	if err != nil {
		return err
	}

	pwHash, pwDoubleHash := computePasswordHash(VAULT_CRED_METHOD_AES_SHA256, password, manager.credentials.Salt)

	encKey, err := encrypted_storage.EncryptFileContents(key, encrypted_storage.AES256_FLAT, pwHash[:])

	if err != nil {
		return err
	}

	manager.credentials.PasswordHash = pwDoubleHash

	manager.credentials.EncryptedKey = encKey

	manager.credentials.TwoFactorAuthEnabled = false
	manager.credentials.TwoFactorAuthEncryptedKey = nil

	return nil
}

// Initializes account credentials (creates a new one if not found)
// user - Root username
// password - Root password
// key - Vault encryption key
// write - Write access
func (manager *VaultCredentialsManager) InitAccountCredentials(user string, password string, key []byte, write bool) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	var accountIndex int = -1

	for i, ac := range manager.credentials.Accounts {
		if ac.User == user {
			accountIndex = i
		}
	}

	if accountIndex == -1 {
		manager.credentials.Accounts = append(manager.credentials.Accounts, VaultCredentialsAccount{
			User: user,
		})
		accountIndex = len(manager.credentials.Accounts) - 1
	}

	account := &manager.credentials.Accounts[accountIndex]

	// Set method
	account.Method = VAULT_CRED_METHOD_AES_SHA256

	// Random salt
	account.Salt = make([]byte, 16)
	_, err := rand.Read(account.Salt)

	if err != nil {
		return err
	}

	pwHash, pwDoubleHash := computePasswordHash(VAULT_CRED_METHOD_AES_SHA256, password, account.Salt)

	encKey, err := encrypted_storage.EncryptFileContents(key, encrypted_storage.AES256_FLAT, pwHash)

	if err != nil {
		return err
	}

	account.EncryptedKey = encKey

	account.PasswordHash = pwDoubleHash

	account.TwoFactorAuthEnabled = false
	account.TwoFactorAuthEncryptedKey = nil

	// Set write access
	account.Write = write

	return nil
}

// Removes an account
// user - Username
func (manager *VaultCredentialsManager) RemoveAccount(user string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	newAccounts := make([]VaultCredentialsAccount, 0)

	for _, account := range manager.credentials.Accounts {
		if account.User != user {
			newAccounts = append(newAccounts, account)
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
		for i := range manager.credentials.Accounts {
			if manager.credentials.Accounts[i].User == user {
				manager.credentials.Accounts[i].User = new_user
				break
			}
		}
	}

	return nil
}

// Changes username
// user - Username
// write - Can write?
func (manager *VaultCredentialsManager) UpdateWritePermission(user string, write bool) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	for i := range manager.credentials.Accounts {
		if manager.credentials.Accounts[i].User == user {
			manager.credentials.Accounts[i].Write = write
			break
		}
	}

	return nil
}

// Changes password of an account
// user - Username
// password - Old account password
// newPassword - New account password
func (manager *VaultCredentialsManager) ChangePassword(user string, password string, newPassword string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	if manager.credentials.User == user {
		// Check old password

		pwHashOld, pwDoubleHashOld := computePasswordHash(manager.credentials.Method, password, manager.credentials.Salt)

		checkResult := subtle.ConstantTimeCompare(pwDoubleHashOld, manager.credentials.PasswordHash) == 1

		if !checkResult {
			return errors.New("invalid password")
		}

		// Decrypt keys

		key, tfaKey, err := decryptCredentialKeys(manager.credentials.EncryptedKey, manager.credentials.TwoFactorAuthEncryptedKey, pwHashOld)

		if err != nil {
			return err
		}

		// Compute new hash

		salt, err := randomSalt()

		if err != nil {
			return err
		}

		pwHash, pwDoubleHash := computePasswordHash(VAULT_CRED_METHOD_AES_SHA256, newPassword, salt)

		// Encrypt keys

		encryptedKey, encryptedTfaKey, err := encryptCredentialKeys(key, tfaKey, pwHash)

		if err != nil {
			return err
		}

		manager.credentials.Method = VAULT_CRED_METHOD_AES_SHA256

		manager.credentials.Salt = salt

		manager.credentials.PasswordHash = pwDoubleHash

		manager.credentials.EncryptedKey = encryptedKey

		manager.credentials.TwoFactorAuthEncryptedKey = encryptedTfaKey
	} else {
		for i := range manager.credentials.Accounts {
			account := &manager.credentials.Accounts[i]

			if account.User == user {
				// Check old password

				pwHashOld, pwDoubleHashOld := computePasswordHash(account.Method, password, account.Salt)

				checkResult := subtle.ConstantTimeCompare(pwDoubleHashOld, account.PasswordHash) == 1

				if !checkResult {
					return errors.New("invalid password")
				}

				// Decrypt keys

				key, tfaKey, err := decryptCredentialKeys(account.EncryptedKey, account.TwoFactorAuthEncryptedKey, pwHashOld)

				if err != nil {
					return err
				}

				// Compute new hash

				salt, err := randomSalt()

				if err != nil {
					return err
				}

				pwHash, pwDoubleHash := computePasswordHash(VAULT_CRED_METHOD_AES_SHA256, newPassword, salt)

				// Encrypt keys

				encryptedKey, encryptedTfaKey, err := encryptCredentialKeys(key, tfaKey, pwHash)

				if err != nil {
					return err
				}

				account.Method = VAULT_CRED_METHOD_AES_SHA256

				account.Salt = salt

				account.PasswordHash = pwDoubleHash

				account.EncryptedKey = encryptedKey

				account.TwoFactorAuthEncryptedKey = encryptedTfaKey

				break
			}
		}
	}

	return nil
}

// Enables two factor authentication for an account
// user - Username
// tfaMethod - TFA method
// tfaKey - TFA key
// password - Account password
func (manager *VaultCredentialsManager) EnableTfa(user string, tfaMethod string, tfaKey []byte, password string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	if manager.credentials.User == user {
		pwHash, pwDoubleHash := computePasswordHash(manager.credentials.Method, password, manager.credentials.Salt)

		checkResult := subtle.ConstantTimeCompare(pwDoubleHash[:], manager.credentials.PasswordHash) == 1

		if !checkResult {
			return errors.New("invalid password")
		}

		encTfaKey, err := encrypted_storage.EncryptFileContents(tfaKey, encrypted_storage.AES256_FLAT, pwHash[:])

		if err != nil {
			return err
		}

		manager.credentials.TwoFactorAuthEnabled = true
		manager.credentials.TwoFactorAuthMethod = tfaMethod

		manager.credentials.TwoFactorAuthEncryptedKey = encTfaKey
	} else {
		for i := range manager.credentials.Accounts {
			account := &manager.credentials.Accounts[i]

			if account.User == user {
				pwHash, pwDoubleHash := computePasswordHash(account.Method, password, account.Salt)

				checkResult := subtle.ConstantTimeCompare(pwDoubleHash[:], account.PasswordHash) == 1

				if !checkResult {
					return errors.New("invalid password")
				}

				encTfaKey, err := encrypted_storage.EncryptFileContents(tfaKey, encrypted_storage.AES256_FLAT, pwHash[:])

				if err != nil {
					return err
				}

				account.TwoFactorAuthEnabled = true
				account.TwoFactorAuthMethod = tfaMethod

				account.TwoFactorAuthEncryptedKey = encTfaKey

				break
			}
		}
	}

	return nil
}

// Disables two factor authentication for an account
// user - Username
// tfaMethod - TFA method
// tfaKey - TFA key
// password - Account password
func (manager *VaultCredentialsManager) DisableTfa(user string) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	if manager.credentials.User == user {
		manager.credentials.TwoFactorAuthEnabled = false
		manager.credentials.TwoFactorAuthMethod = ""

		manager.credentials.TwoFactorAuthEncryptedKey = nil
	} else {
		for i := range manager.credentials.Accounts {
			account := &manager.credentials.Accounts[i]

			if account.User == user {
				account.TwoFactorAuthEnabled = false
				account.TwoFactorAuthMethod = ""

				account.TwoFactorAuthEncryptedKey = nil

				break
			}
		}
	}

	return nil
}

// Disables two factor authentication for an account
// user - Username
// authConfirmationEnabled - True if auth confirmation is enabled
// authConfirmationMethod - Auth confirmation method
// authConfirmationPeriodSeconds - Auth confirmation period (seconds)
func (manager *VaultCredentialsManager) ChangeSecuritySettings(user string, authConfirmationEnabled bool, authConfirmationMethod string, authConfirmationPeriodSeconds uint32) error {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	// Set user
	if manager.credentials.User == user {
		manager.credentials.AuthConfirmationEnabled = &authConfirmationEnabled
		manager.credentials.AuthConfirmationMethod = authConfirmationMethod
		manager.credentials.AuthConfirmationPeriod = &authConfirmationPeriodSeconds
	} else {
		for i := range manager.credentials.Accounts {
			account := &manager.credentials.Accounts[i]

			if account.User == user {
				account.AuthConfirmationEnabled = &authConfirmationEnabled
				account.AuthConfirmationMethod = authConfirmationMethod
				account.AuthConfirmationPeriod = &authConfirmationPeriodSeconds

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
	err = os.WriteFile(tFile, jsonData, FILE_PERMISSION)
	if err != nil {
		return err
	}

	// Move to the original path
	err = RenameAndReplace(tFile, manager.file)
	if err != nil {
		return err
	}

	return nil
}

// Credentials check result
type CredentialsCheckResult struct {
	root  bool // Root account
	write bool // Write access

	tfa       bool   // TFA enabled?
	tfaMethod string // TFS method

	authConfirm       bool          // Auth confirmation enabled?
	authConfirmMethod string        // Auth confirmation method
	authConfirmPeriod time.Duration // Auth confirmation period
}

// Find the account credential details
func (manager *VaultCredentialsManager) getAccount(user string) (found bool, result *CredentialsCheckResult, pwMethod string, pwSalt []byte, pwExpectedHash []byte, pwEncryptedKey []byte, encTfaKey []byte) {
	result = &CredentialsCheckResult{
		root:  false,
		write: false,
	}

	if manager.credentials.User == user {
		result.root = true
		result.write = true

		result.tfa = manager.credentials.TwoFactorAuthEnabled
		result.tfaMethod = manager.credentials.TwoFactorAuthMethod

		result.authConfirm = GetAuthConfirmationEnabled(manager.credentials.AuthConfirmationEnabled)
		result.authConfirmMethod = manager.credentials.AuthConfirmationMethod
		result.authConfirmPeriod = GetAuthConfirmationPeriod(manager.credentials.AuthConfirmationPeriod)

		pwMethod = manager.credentials.Method
		pwSalt = manager.credentials.Salt
		pwExpectedHash = manager.credentials.PasswordHash
		pwEncryptedKey = manager.credentials.EncryptedKey
		encTfaKey = manager.credentials.TwoFactorAuthEncryptedKey
	} else {
		result.root = false

		var foundAccount bool = false

		for i := 0; i < len(manager.credentials.Accounts); i++ {
			account := &manager.credentials.Accounts[i]
			if account.User == user {
				foundAccount = true

				result.tfa = account.TwoFactorAuthEnabled
				result.tfaMethod = account.TwoFactorAuthMethod

				result.authConfirm = GetAuthConfirmationEnabled(account.AuthConfirmationEnabled)
				result.authConfirmMethod = account.AuthConfirmationMethod
				result.authConfirmPeriod = GetAuthConfirmationPeriod(account.AuthConfirmationPeriod)

				result.write = account.Write
				pwMethod = account.Method
				pwSalt = account.Salt
				pwExpectedHash = account.PasswordHash
				pwEncryptedKey = account.EncryptedKey
				encTfaKey = account.TwoFactorAuthEncryptedKey
				break
			}
		}

		if !foundAccount {
			return false, nil, "", nil, nil, nil, nil
		}
	}

	return true, result, pwMethod, pwSalt, pwExpectedHash, pwEncryptedKey, encTfaKey
}

// Checks password
// user - Username
// password - Password
// Returns (success, result)
// - success is true if the credentials check succeeded
// - result contains info if the credentials are root credentials and if the have write access, also it it has tfa enabled
func (manager *VaultCredentialsManager) CheckPassword(user string, password string) (bool, *CredentialsCheckResult) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	found, result, pwMethod, pwSalt, pwExpectedHash, _, _ := manager.getAccount(user)

	if !found {
		return false, nil
	}

	pwHash, pwDoubleHash := computePasswordHash(pwMethod, password, pwSalt)

	if pwHash == nil || pwDoubleHash == nil {
		return false, nil
	}

	checkResult := subtle.ConstantTimeCompare(pwDoubleHash[:], pwExpectedHash) == 1

	if checkResult {
		return true, result
	} else {
		return false, nil
	}
}

// Unlocks vaults and gets decryption key
// user - Username
// password - Password
// Returns the key and the result (root and write info)
// If fails, it returns an error
func (manager *VaultCredentialsManager) UnlockVault(user string, password string) (key []byte, checkRes *CredentialsCheckResult, tfaKey []byte, err error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	found, result, pwMethod, pwSalt, pwExpectedHash, pwEncryptedKey, encTfaKey := manager.getAccount(user)

	if !found {
		return nil, nil, nil, errors.New("unknown user")
	}

	pwHash, pwDoubleHash := computePasswordHash(pwMethod, password, pwSalt)

	if pwHash == nil || pwDoubleHash == nil {
		return nil, nil, nil, errors.New("unknown credentials method")
	}

	if subtle.ConstantTimeCompare(pwDoubleHash[:], pwExpectedHash) != 1 {
		return nil, nil, nil, errors.New("invalid credentials")
	}

	key, tfaKey, err = decryptCredentialKeys(pwEncryptedKey, encTfaKey, pwHash)

	if err != nil {
		return nil, nil, nil, err
	}

	return key, result, tfaKey, nil
}

// Gets vault fingerprint
func (manager *VaultCredentialsManager) GetFingerprint() string {
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

	result := make([]ApiAdminAccountEntry, len(manager.credentials.Accounts))

	for i := 0; i < len(manager.credentials.Accounts); i++ {
		result[i] = ApiAdminAccountEntry{
			Username: manager.credentials.Accounts[i].User,
			Write:    manager.credentials.Accounts[i].Write,
		}
	}

	return result
}
