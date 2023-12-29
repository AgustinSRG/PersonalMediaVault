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

	write bool // Write permissions
	root  bool // Root permission

	key []byte // Decryption key

	not_after int64 // Expiration
}

// Session Manager
type SessionManager struct {
	vault *Vault // Reference to vault

	lock *sync.Mutex // Mutex to access the data

	next_index uint64

	sessions map[string]*ActiveSession // Sessions
}

// Initialization
func (sm *SessionManager) Initialize(vault *Vault) {
	sm.vault = vault
	sm.lock = &sync.Mutex{}
	sm.next_index = 0
	sm.sessions = make(map[string]*ActiveSession)

	go sm.RunSessionChecker()
}

// Creates a session
// user - Username
// key - Vault decryption key
// root - Root access
// write - Write access
// expirationTime - Expiration time (Milliseconds)
// Returns an error if failed, and the session ID if successful
func (sm *SessionManager) CreateSession(user string, key []byte, root bool, write bool, expirationTime int64) (error, string) {
	sessionBytes := make([]byte, 32)
	_, err_rand := rand.Read(sessionBytes)

	if err_rand != nil {
		return err_rand, ""
	}

	sessionId := hex.EncodeToString(sessionBytes)

	sm.lock.Lock()

	isFirstSession := len(sm.sessions) == 0

	newSession := ActiveSession{
		index:     sm.next_index,
		id:        sessionId,
		user:      user,
		key:       key,
		write:     write,
		root:      root,
		not_after: time.Now().UnixMilli() + expirationTime,
	}

	sm.next_index++

	sm.sessions[sessionId] = &newSession

	sm.lock.Unlock()

	// Call task manager to start pending tasks
	err := sm.vault.tasks.OnNewSession(&newSession)

	if err != nil {
		LogError(err)
	}

	if isFirstSession {
		// Pre-cache tags and albums
		go sm.vault.tags.PreCacheTags(key)
		go sm.vault.albums.PreCacheAlbums(key)
	}

	return nil, sessionId
}

// Closes a session
// session_id - Session token
// Returns true only if the session existed
func (sm *SessionManager) CloseSession(session_id string) bool {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if sm.sessions[session_id] != nil {
		delete(sm.sessions, session_id)
		return true
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

// Clears expired session
func (sm *SessionManager) ClearExpiredSessions() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for sid, session := range sm.sessions {
		now := time.Now().UnixMilli()
		if session.not_after < now {
			delete(sm.sessions, sid)
		}
	}
}

// Finds a session
// session_id - Session token
// Returns the session or nil if not found
func (sm *SessionManager) FindSession(session_id string) *ActiveSession {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for sid, session := range sm.sessions {
		now := time.Now().UnixMilli()
		if session.not_after < now {
			continue // Expired session
		}

		if subtle.ConstantTimeCompare([]byte(sid), []byte(session_id)) == 1 {
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

	for _, session := range sm.sessions {
		return session
	}

	return nil
}

// Changes session username
// user - Old username
// new_user - New username
func (sm *SessionManager) ChangeUsername(user string, new_user string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for _, session := range sm.sessions {
		if session.user == user {
			session.user = new_user
		}
	}
}

// Removes all the sessions for an user
// user - Username
func (sm *SessionManager) RemoveUserSessions(user string) {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	for sid, session := range sm.sessions {
		if session.user == user {
			delete(sm.sessions, sid)
		}
	}
}
