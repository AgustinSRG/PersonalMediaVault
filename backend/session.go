// Session manager

package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"sync"
	"time"
)

const SESSION_EXPIRATION_TIME = 24 * 60 * 60 * 1000

// User session
type ActiveSession struct {
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

	sessions map[string]*ActiveSession // Sessions
}

// Initialization
func (sm *SessionManager) Initialize(vault *Vault) {
	sm.vault = vault
	sm.lock = &sync.Mutex{}
	sm.sessions = make(map[string]*ActiveSession)

	go sm.RunSessionChecker()
}

// Creates a session
// user - Username
// key - Vault decryption key
// root - Root access
// write - Write access
// Returns the session ID
func (sm *SessionManager) CreateSession(user string, key []byte, root bool, write bool) string {
	sessionBytes := make([]byte, 32)
	rand.Read(sessionBytes)
	sessionId := hex.EncodeToString(sessionBytes)

	sm.lock.Lock()

	newSession := ActiveSession{
		id:        sessionId,
		user:      user,
		key:       key,
		write:     write,
		root:      root,
		not_after: time.Now().UnixMilli() + SESSION_EXPIRATION_TIME,
	}

	sm.sessions[sessionId] = &newSession

	sm.lock.Unlock()

	// Call task manager to start pending tasks
	sm.vault.tasks.OnNewSession(&newSession)

	return sessionId
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
	for true {
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
