// Session manager

package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
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

	user string // User

	invitedBy string // Invited by

	write bool // Write permissions
	root  bool // Root permission

	key []byte // Decryption key

	timestamp int64 // Creation timestamp
	not_after int64 // Expiration

	// Below: Auth confirmation properties

	tfa       bool   // Two factor auth enabled?
	tfaKey    []byte // Two factor auth key
	tfaMethod string // Two factor auth method

	authConfirmationEnabled bool          // Auth confirmation enabled?
	authConfirmationMethod  string        // Auth confirmation method (tfa or pw)
	authConfirmationPeriod  time.Duration // Auth confirmation period

	authConfirmationLastTime int64 // Last time auth confirmation was successfully (Unix Millis)
	authConfirmationLastTry  int64 // Last time auth confirmation wsa tried and failed (Unix Millis)

	authConfirmationMutex *sync.Mutex // Mutex for auth confirmation properties
}

// Session Manager
type SessionManager struct {
	vault *Vault // Reference to vault

	lock *sync.Mutex // Mutex to access the data

	next_index uint64 // Counter to make unique session indexes

	sessions []([]*ActiveSession) // Sessions
}

// Initialization
func (sm *SessionManager) Initialize(vault *Vault) {
	sm.vault = vault
	sm.lock = &sync.Mutex{}
	sm.next_index = 0
	sm.sessions = make([]([]*ActiveSession), 256)

	for i := 0; i < 256; i++ {
		sm.sessions[i] = make([]*ActiveSession, 0)
	}

	go sm.RunSessionChecker()
}

// Options to create a session
type CreateSessionOptions struct {
	user string // Username

	key []byte // Vault key

	root  bool // Is root user?
	write bool // Has write access?

	expirationTime int64 // Expiration time (Milliseconds)

	invitedBy string // User who invited

	tfa       bool   // Two factor auth enabled?
	tfaKey    []byte // Two factor auth key
	tfaMethod string // Two factor auth method

	authConfirmationEnabled bool          // Auth confirmation enabled?
	authConfirmationMethod  string        // Auth confirmation method (tfa or pw)
	authConfirmationPeriod  time.Duration // Auth confirmation period
}

// Creates a session
// user - Username
// key - Vault decryption key
// root - Root access
// write - Write access
// expirationTime - Expiration time (Milliseconds)
// invitedBy - User who invited
// Returns an error if failed, and the session ID if successful
func (sm *SessionManager) CreateSession(options CreateSessionOptions) (string, error) {
	sessionBytes := make([]byte, 32)
	_, err_rand := rand.Read(sessionBytes)

	if err_rand != nil {
		return "", err_rand
	}

	sessionHash := uint8(sessionBytes[0])
	sessionId := hex.EncodeToString(sessionBytes)

	sm.lock.Lock()

	isFirstSession := len(sm.sessions) == 0

	now := time.Now().UnixMilli()

	newSession := ActiveSession{
		index:                    sm.next_index,
		id:                       sessionId,
		user:                     options.user,
		invitedBy:                options.invitedBy,
		key:                      options.key,
		write:                    options.write,
		root:                     options.root,
		timestamp:                now,
		not_after:                now + options.expirationTime,
		tfa:                      options.tfa,
		tfaKey:                   options.tfaKey,
		tfaMethod:                options.tfaMethod,
		authConfirmationEnabled:  options.authConfirmationEnabled,
		authConfirmationMethod:   options.authConfirmationMethod,
		authConfirmationPeriod:   options.authConfirmationPeriod,
		authConfirmationLastTime: 0,
		authConfirmationLastTry:  0,
		authConfirmationMutex:    &sync.Mutex{},
	}

	sm.next_index++

	sm.sessions[sessionHash] = append(sm.sessions[sessionHash], &newSession)

	sm.lock.Unlock()

	// Call task manager to start pending tasks
	err := sm.vault.tasks.OnNewSession(&newSession)

	if err != nil {
		LogError(err)
	}

	if isFirstSession {
		// Pre-cache tags and albums
		go sm.vault.tags.PreCacheTags(options.key)
		go sm.vault.albums.PreCacheAlbums(options.key)
	}

	return sessionId, nil
}

// Closes a session
// session_id - Session token
// Returns true only if the session existed
func (sm *SessionManager) CloseSession(session_id string) bool {
	sessionHash := getSessionIdHash(session_id)

	sm.lock.Lock()
	defer sm.lock.Unlock()

	sessionsList := sm.sessions[sessionHash]

	for i := 0; i < len(sessionsList); i++ {
		if sessionsList[i].id == session_id {
			sessionsList[i] = sessionsList[len(sessionsList)-1]
			sm.sessions[sessionHash] = sessionsList[:len(sessionsList)-1]
			return true
		}
	}

	return false
}

// Clear expired sessions, check once each 5 minutes
func (sm *SessionManager) RunSessionChecker() {
	for {
		time.Sleep(5 * time.Minute)

		sm.ClearExpiredSessions()
	}
}

// Clears expired sessions
func (sm *SessionManager) ClearExpiredSessions() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	now := time.Now().UnixMilli()

	for i := 0; i < len(sm.sessions); i++ {
		sessionsList := sm.sessions[i]

		for j := 0; j < len(sessionsList); j++ {
			if sessionsList[j].not_after < now {
				// Delete
				sessionsList[j] = sessionsList[len(sessionsList)-1]
				sessionsList = sessionsList[:len(sessionsList)-1]
				sm.sessions[i] = sessionsList
				j--
			}
		}
	}
}

// Finds a session
// session_id - Session token
// Returns the session or nil if not found
func (sm *SessionManager) FindSession(session_id string) *ActiveSession {
	sessionHash := getSessionIdHash(session_id)

	sm.lock.Lock()
	defer sm.lock.Unlock()

	sessionsList := sm.sessions[sessionHash]
	now := time.Now().UnixMilli()

	for i := 0; i < len(sessionsList); i++ {
		session := sessionsList[i]

		if session.not_after < now {
			continue // Expired session
		}

		if subtle.ConstantTimeCompare([]byte(session.id), []byte(session_id)) == 1 {
			return session
		}
	}

	return nil
}

// Finds any session
// Returns a session, or nil if the vault is locked
func (sm *SessionManager) FindAnySession() *ActiveSession {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := 0; i < len(sm.sessions); i++ {
		sessionsList := sm.sessions[i]

		if len(sessionsList) > 0 {
			return sessionsList[0]
		}
	}

	return nil
}

// Changes session username
// user - Old username
// new_user - New username
func (sm *SessionManager) ChangeUsername(user string, new_user string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := 0; i < len(sm.sessions); i++ {
		sessionsList := sm.sessions[i]

		for j := 0; j < len(sessionsList); j++ {
			session := sessionsList[j]

			if session.user == user {
				session.user = new_user
			}

			if session.invitedBy == user {
				session.invitedBy = new_user
			}
		}
	}
}

// Removes all the sessions for an user
// user - Username
func (sm *SessionManager) RemoveUserSessions(user string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := 0; j < len(sessionsList); j++ {
			if sessionsList[j].user == user || sessionsList[j].invitedBy == user {
				// Delete
				sessionsList[j] = sessionsList[len(sessionsList)-1]
				sessionsList = sessionsList[:len(sessionsList)-1]
				sm.sessions[i] = sessionsList
				j--
			}
		}
	}
}

// Updates write permissions of all sessions of user
// user - Username
// write - Write permission
func (sm *SessionManager) UpdateUserSessions(user string, write bool) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			if sessionsList[j].user == user {
				// Update
				sessionsList[j].write = write
			}
		}
	}
}

// Updates auth confirmation details of all sessions of user
// user - Username
// authConfirmationEnabled - True if auth confirmation is enabled
// authConfirmationMethod - Auth confirmation method
// authConfirmationPeriodSeconds - Auth confirmation period (seconds)
func (sm *SessionManager) UpdateUserSessionsAuthConfirmation(user string, authConfirmationEnabled bool, authConfirmationMethod string, authConfirmationPeriodSeconds uint32) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			if sessionsList[j].user == user {
				// Update
				sessionsList[j].authConfirmationEnabled = authConfirmationEnabled
				sessionsList[j].authConfirmationMethod = authConfirmationMethod
				sessionsList[j].authConfirmationPeriod = time.Duration(authConfirmationPeriodSeconds) * time.Second
			}
		}
	}
}

// Enables two factor authentication of all sessions of user
// user - Username
// tfaKey - Two factor auth key
// tfaMethod - Two factor auth method
func (sm *SessionManager) UpdateUserSessionsEnableTfa(user string, tfaKey []byte, tfaMethod string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			if sessionsList[j].user == user {
				// Update
				sessionsList[j].tfa = true
				sessionsList[j].tfaKey = tfaKey
				sessionsList[j].tfaMethod = tfaMethod
			}
		}
	}
}

// Disables two factor authentication of all sessions of user
// user - Username
func (sm *SessionManager) UpdateUserSessionsDisableTfa(user string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			if sessionsList[j].user == user {
				// Update
				sessionsList[j].tfa = false
				sessionsList[j].tfaKey = nil
				sessionsList[j].tfaMethod = ""
			}
		}
	}
}

// Removes an invite session
// invitedBy - User who invited
// index - Session unique index
func (sm *SessionManager) RemoveInviteSession(invitedBy string, index uint64) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			if sessionsList[j].invitedBy == invitedBy && sessionsList[j].index == index {
				// Delete
				sessionsList[j] = sessionsList[len(sessionsList)-1]
				sessionsList = sessionsList[:len(sessionsList)-1]
				sm.sessions[i] = sessionsList
				return
			}
		}
	}
}

// Finds sessions invited by an user
// user - The user
// Returns the list of sessions
func (sm *SessionManager) FindInviteSessions(user string) []InviteCodeSessionItem {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	result := make([]InviteCodeSessionItem, 0)

	for i := range sm.sessions {
		sessionsList := sm.sessions[i]

		for j := range sessionsList {
			session := sessionsList[j]

			if session.user == "" && session.invitedBy == user {
				result = append(result, InviteCodeSessionItem{
					Index:      session.index,
					Timestamp:  session.timestamp,
					Expiration: session.not_after,
				})
			}
		}
	}

	return result
}

// Gets two factor authentication properties to check codes
func (s *ActiveSession) GetTwoFactorAuth() (tfaEnabled bool, tfaMethod string, tfaKey []byte) {
	s.authConfirmationMutex.Lock()
	defer s.authConfirmationMutex.Unlock()

	return s.tfa, s.tfaMethod, s.tfaKey
}

// Checks auth confirmation
func (s *ActiveSession) CheckAuthConfirmation(code string, force bool) (success bool, cooldown bool, requiredMethod string) {
	now := time.Now().UnixMilli()

	s.authConfirmationMutex.Lock()
	defer s.authConfirmationMutex.Unlock()

	if !s.authConfirmationEnabled {
		return true, false, "" // No auth confirmation
	}

	if now-s.authConfirmationLastTry < AUTH_FAIL_COOLDOWN {
		return false, true, s.authConfirmationMethod
	}

	if !force {
		if time.Duration(now-s.authConfirmationLastTime)*time.Millisecond < s.authConfirmationPeriod {
			// No need to check, recently confirmed
			return true, false, s.authConfirmationMethod
		}
	}

	if code == "" {
		return false, false, s.authConfirmationMethod
	}

	if s.authConfirmationMethod == "pw" {
		// Password check
		valid, _ := GetVault().credentials.CheckPassword(s.user, code)

		if !valid {
			// Set last try
			s.authConfirmationLastTry = now
		}

		return valid, false, "pw"
	} else {
		// TFA check

		valid := validateTwoFactorAuthCode(s.tfaMethod, s.tfaKey, code)

		if !valid {
			// Set last try
			s.authConfirmationLastTry = now
		}

		return valid, false, "tfa"
	}
}

// Computes session hash for the sessions hash table
// sessionId - The session id
// Returns a number from 0 to 255
func getSessionIdHash(sessionId string) uint8 {
	if len(sessionId) < 2 {
		return 0
	}

	prefix := sessionId[0:2]

	dec, err := hex.DecodeString(prefix)

	if err != nil || len(dec) < 1 {
		return 0
	}

	return dec[0]
}
