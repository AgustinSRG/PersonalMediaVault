// Credentials manager

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"os"
)

const VAULT_CRED_METHOD_AES_SHA256 string = "aes256/sha256/salt16"

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

	VaultFingerprint string `json:"fingerprint"` // Vault fingerprint, user to distinguish between vaults

	Accounts []VaultCredentialsAccount `json:"accounts"` // Accounts
}

// Generates and return a random key for encryption
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		return nil, err
	}

	return key, nil
}

// Creates a credentials set
// Params:
//
//	user - Username
//	password - Password
//	fingerprint - Vault fingerprint
//
// Returns:
//
//	ex - Error
//	credentials - The new credentials set
func MakeCredentials(user string, password string, fingerprint string) (credentials *VaultCredentials, ex error) {
	// Generate random key
	key, err := GenerateRandomKey()

	if err != nil {
		return nil, err
	}

	// Generate random salt
	randomSalt := make([]byte, 16)
	_, err = rand.Read(randomSalt)

	if err != nil {
		return nil, err
	}

	// Generate encrypted key
	pwBytes := []byte(password)
	ctBytes := make([]byte, len(pwBytes)+16)
	copy(ctBytes[:len(pwBytes)], pwBytes)
	copy(ctBytes[len(pwBytes):], randomSalt)
	pwHash := sha256.Sum256(ctBytes)

	encKey, err := encryptFileContents(key, AES256_FLAT, pwHash[:])

	if err != nil {
		return nil, err
	}

	// Password hash

	pwDoubleHash := sha256.Sum256(pwHash[:])

	// Return

	var newCredentials VaultCredentials

	newCredentials.Method = VAULT_CRED_METHOD_AES_SHA256
	newCredentials.User = user
	newCredentials.VaultFingerprint = fingerprint
	newCredentials.PasswordHash = pwDoubleHash[:]
	newCredentials.Salt = randomSalt
	newCredentials.Accounts = make([]VaultCredentialsAccount, 0)
	newCredentials.EncryptedKey = encKey

	return &newCredentials, nil
}

// Checks password
// Params:
//
// password - The password to check
// method - Credentials set method
// passwordHash - Password hash
// salt - Random salt
//
// Returns: True only if the password is correct
func CheckPassword(password string, method string, passwordHash []byte, salt []byte) bool {
	if method == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], salt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		checkResult := subtle.ConstantTimeCompare(pwDoubleHash[:], passwordHash) == 1

		if checkResult {
			return true
		} else {
			return false
		}
	} else {
		// Unknown method
		return false
	}
}

// Decrypts key
// Params:
//
//	password - The password to check
//	method - Credentials set method
//	passwordHash - Password hash
//	salt - Random salt
//	encryptedKey - Encrypted key
//
// Returns:
//
//	decryptedKey - Decrypted key
//	ex - Error
func DecryptKey(password string, method string, passwordHash []byte, salt []byte, encryptedKey []byte) (decryptedKey []byte, ex error) {
	if method == VAULT_CRED_METHOD_AES_SHA256 {
		// Compute password hash
		pwBytes := []byte(password)
		ctBytes := make([]byte, len(pwBytes)+16)
		copy(ctBytes[:len(pwBytes)], pwBytes)
		copy(ctBytes[len(pwBytes):], salt)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		if subtle.ConstantTimeCompare(pwDoubleHash[:], passwordHash) != 1 {
			return nil, errors.New("Invalid credentials")
		}

		// Decrypt key
		key, err := decryptFileContents(encryptedKey, pwHash[:])

		if err != nil {
			return nil, err
		}

		return key, nil
	} else {
		return nil, errors.New("Unknown credentials method")
	}
}

func ReadVaultCredentials(file string) (*VaultCredentials, error) {
	// Read file
	b, err := os.ReadFile(file)

	if err != nil {
		return nil, err
	}

	// Parse
	var credentials VaultCredentials

	err = json.Unmarshal(b, &credentials)

	if err != nil {
		return nil, err
	}

	return &credentials, nil
}

func (vc *VaultCredentials) WriteToFile(file string, tmpFile string) error {
	// Get the json data
	jsonData, err := json.Marshal(vc)

	if err != nil {
		return err
	}

	// Write file
	err = os.WriteFile(tmpFile, jsonData, FILE_PERMISSION)
	if err != nil {
		return err
	}

	// Move to the original path
	err = RenameAndReplace(tmpFile, file)
	if err != nil {
		os.Remove(tmpFile)
		return err
	}

	return nil
}
