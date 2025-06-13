// Session manager

package main

import (
	"sync"
	"time"
)

const SESSION_EXPIRATION_TIME_DAY = 24 * 60 * 60 * 1000
const SESSION_EXPIRATION_TIME_WEEK = 7 * SESSION_EXPIRATION_TIME_DAY
const SESSION_EXPIRATION_TIME_MONTH = 30 * SESSION_EXPIRATION_TIME_DAY
const SESSION_EXPIRATION_TIME_YEAR = 365 * SESSION_EXPIRATION_TIME_DAY

// User session
type ActiveSession struct {
	index uint64 // Session index

	id string // Session token

	mu *sync.Mutex // Mutex for the struct

	user string // User

	invitedBy string // Invited by

	root  bool // Root permission
	write bool // Write permissions

	key []byte // Decryption key

	timestamp int64 // Creation timestamp
	not_after int64 // Expiration

	tfa       bool   // Two factor auth enabled?
	tfaKey    []byte // Two factor auth key
	tfaMethod string // Two factor auth method

	authConfirmationEnabled bool          // Auth confirmation enabled?
	authConfirmationMethod  string        // Auth confirmation method (tfa or pw)
	authConfirmationPeriod  time.Duration // Auth confirmation period

	authConfirmationLastTime int64 // Last time auth confirmation was successfully (Unix Millis)
	authConfirmationLastTry  int64 // Last time auth confirmation wsa tried and failed (Unix Millis)
}

// Gets the user for this session
func (s *ActiveSession) GetUser() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.user
}

// Checks if the session is an user
// Returns true if is an user, false if invited
func (s *ActiveSession) IsUser() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.user) > 0
}

// Checks if the session has write permissions
func (s *ActiveSession) CanWrite() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.write
}

// Gets details for auth context API
func (s *ActiveSession) GetContextDetails() (username string, root bool, write bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.user, s.root, s.write
}

// Gets two factor authentication properties to check codes
func (s *ActiveSession) GetTwoFactorAuth() (tfaEnabled bool, tfaMethod string, tfaKey []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tfa, s.tfaMethod, s.tfaKey
}

// Checks auth confirmation
func (s *ActiveSession) CheckAuthConfirmation(password string, tfaCode string, onlyTfa bool) (success bool, cooldown bool, requiredMethod string) {
	now := time.Now().UnixMilli()

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.authConfirmationEnabled {
		return true, false, "" // No auth confirmation
	}

	if now-s.authConfirmationLastTry < AUTH_FAIL_COOLDOWN {
		return false, true, s.authConfirmationMethod
	}

	if time.Duration(now-s.authConfirmationLastTime)*time.Millisecond < s.authConfirmationPeriod {
		// No need to check, recently confirmed
		return true, false, s.authConfirmationMethod
	}

	if s.authConfirmationMethod == "pw" || !s.tfa {
		if onlyTfa {
			return true, false, "pw"
		}

		if password == "" {
			return false, false, "pw"
		}

		// Password check
		valid, _ := GetVault().credentials.CheckPassword(s.user, password)

		if !valid {
			// Set last try
			s.authConfirmationLastTry = now
		} else {
			s.authConfirmationLastTime = now
		}

		return valid, false, "pw"
	} else {
		if tfaCode == "" {
			return false, false, "tfa"
		}

		// TFA check

		valid := validateTwoFactorAuthCode(s.tfaMethod, s.tfaKey, tfaCode)

		if !valid {
			// Set last try
			s.authConfirmationLastTry = now
		} else {
			s.authConfirmationLastTime = now
		}

		return valid, false, "tfa"
	}
}

// Gets account security options for the API
func (s *ActiveSession) GetAccountSecurityOptions() *AccountSecurityOptions {
	s.mu.Lock()
	defer s.mu.Unlock()

	authConfirmMethod := s.authConfirmationMethod

	if authConfirmMethod == "" {
		authConfirmMethod = "tfa"
	}

	return &AccountSecurityOptions{
		TwoFactorAuthEnabled:          s.tfa,
		TwoFactorAuthMethod:           s.tfaMethod,
		AuthConfirmationEnabled:       s.authConfirmationEnabled,
		AuthConfirmationMethod:        authConfirmMethod,
		AuthConfirmationPeriodSeconds: uint32(s.authConfirmationPeriod / time.Second),
	}
}
