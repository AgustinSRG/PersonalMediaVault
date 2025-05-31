// Utility functions for credentials

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

// Method: SHA-256 hash + AES-256 encryption
const VAULT_CRED_METHOD_AES_SHA256 string = "aes256/sha256/salt16"

// Size of salt (bytes)
const SALT_LENGTH_BYTES = 16

// Creates a random salt to compute password hash
func randomSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)

	if err != nil {
		return nil, err
	}

	return salt, nil
}

// Computes password hash
// method - The hash method
// password - The password
// salt - The salt
// Returns the password hash (used for encryption) and the double hash (used for checking)
func computePasswordHash(method string, password string, salt []byte) (pwHash []byte, pwDoubleHash []byte) {
	if method == VAULT_CRED_METHOD_AES_SHA256 {
		pwBytes := []byte(password)
		ctBytes := append(pwBytes, salt...)
		pwHash := sha256.Sum256(ctBytes)
		pwDoubleHash := sha256.Sum256(pwHash[:])

		return pwHash[:], pwDoubleHash[:]
	} else {
		return nil, nil
	}
}

// Encrypts the credentials keys using the password hash
// key - The vault encryption key
// tfaKey - The two factor auth key
// pwHash - The password hash
// Returns the encrypted vault key and the encrypted tfa key
func encryptCredentialKeys(key []byte, tfaKey []byte, pwHash []byte) (encryptedKey []byte, encryptedTfaKey []byte, err error) {
	encryptedKey, err = encrypted_storage.EncryptFileContents(key, encrypted_storage.AES256_FLAT, pwHash)

	if err != nil {
		return nil, nil, err
	}

	if len(tfaKey) > 0 {
		encryptedTfaKey, err = encrypted_storage.EncryptFileContents(tfaKey, encrypted_storage.AES256_FLAT, pwHash)

		if err != nil {
			return nil, nil, err
		}
	} else {
		encryptedTfaKey = nil
	}

	return encryptedKey, encryptedTfaKey, nil
}

// Decrypts encrypted credentials keys with the password hash
// encryptedKey - Encrypted vault key
// encryptedTfaKey - Encrypted tfa key
// pwHash - The password hash
// Returns the vault key and the tfa key
func decryptCredentialKeys(encryptedKey []byte, encryptedTfaKey []byte, pwHash []byte) (key []byte, tfaKey []byte, err error) {
	// Decrypt key
	key, err = encrypted_storage.DecryptFileContents(encryptedKey, pwHash)

	if err != nil {
		return nil, nil, err
	}

	// Decrypt tfa key (if available)

	if len(encryptedTfaKey) > 0 {
		tfaKey, err = encrypted_storage.DecryptFileContents(encryptedTfaKey, pwHash)

		if err != nil {
			return nil, nil, err
		}
	} else {
		tfaKey = nil
	}

	return key, tfaKey, nil
}

type TimeOtpOptions struct {
	// The period (seconds)
	Period uint

	// Algorithm to use for HMAC.
	Algorithm otp.Algorithm

	// Clock skew allowed?
	AllowClockSkew bool
}

// Gets skeps in period counts
func (o *TimeOtpOptions) Skew() uint {
	if o.AllowClockSkew {
		return 1
	} else {
		return 0
	}
}

// Turns the options into a method string
func (o *TimeOtpOptions) ToMethodString() string {
	algo := "sha1"

	switch o.Algorithm {
	case otp.AlgorithmSHA256:
		algo = "sha256"
	case otp.AlgorithmSHA512:
		algo = "sha512"
	}

	clockSkewFlag := "1"

	if !o.AllowClockSkew {
		clockSkewFlag = "0"
	}

	return "totp:" + algo + ":" + fmt.Sprint(o.Period) + ":" + clockSkewFlag
}

// Parses time otp options from method string
// method - The method string
func parseTimeOtpOptions(method string) (*TimeOtpOptions, error) {
	parts := strings.Split(method, ":")

	if len(parts) != 4 {
		return nil, errors.New("malformed method string: expected 4 parts")
	}

	if parts[0] != "totp" {
		return nil, errors.New("malformed method string: expected totp as the first part")
	}

	var algo otp.Algorithm

	switch parts[1] {
	case "sha1":
		algo = otp.AlgorithmSHA1
	case "sha256":
		algo = otp.AlgorithmSHA256
	case "sha512":
		algo = otp.AlgorithmSHA512
	default:
		return nil, errors.New("invalid hmac algorithm")
	}

	p, err := strconv.ParseUint(parts[2], 10, 32)

	if err != nil {
		return nil, err
	}

	allowClockSkew := parts[3] == "1"

	return &TimeOtpOptions{
		Algorithm:      algo,
		Period:         uint(p),
		AllowClockSkew: allowClockSkew,
	}, nil
}

func generateTimeOtpKey(issuer string, user string, options *TimeOtpOptions) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: user,
		Period:      options.Period,
		Algorithm:   options.Algorithm,
		Digits:      otp.DigitsSix,
	})

	if err != nil {
		return nil, err
	}

	return key, nil
}

// Validates TOTP code
func validateTimeOtpAuthCode(method string, key []byte, code string) bool {
	options, err := parseTimeOtpOptions(method)

	if err != nil {
		LogDebug("Error: " + err.Error())
		return false
	}

	if err != nil {
		LogDebug("Error: " + err.Error())
		return false
	}

	valid, err := totp.ValidateCustom(code, string(key), time.Now(), totp.ValidateOpts{
		Period:    options.Period,
		Algorithm: options.Algorithm,
		Digits:    otp.DigitsSix,
		Skew:      options.Skew(),
	})

	if err != nil {
		LogDebug("Error: " + err.Error())
		return false
	}

	return valid
}

// Validates TFA code
func validateTwoFactorAuthCode(method string, key []byte, code string) bool {
	parts := strings.Split(method, ":")

	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "totp":
		return validateTimeOtpAuthCode(method, key, code)
	default:
		LogDebug("TFA: Unrecognized method")
		return false
	}
}
