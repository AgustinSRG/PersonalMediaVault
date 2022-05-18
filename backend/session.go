// Session manager

package main

import (
	"crypto/subtle"
	"encoding/hex"
	"math/rand"
	"sync"
	"time"
)

const SESSION_EXPIRATION_TIME = 24 * 60 * 60 * 1000

type ActiveSession struct {
	user      string
	key       []byte
	not_after int64
}

type SessionManager struct {
	vault    *Vault
	lock     *sync.Mutex
	sessions map[string]*ActiveSession
}

func (sm *SessionManager) Initialize(vault *Vault) {
	sm.vault = vault
	sm.lock = &sync.Mutex{}
	sm.sessions = make(map[string]*ActiveSession)

	go sm.RunSessionChecker()
}

func (sm *SessionManager) CreateSession(user string, key []byte) string {
	sessionBytes := make([]byte, 32)
	rand.Read(sessionBytes)
	sessionId := hex.EncodeToString(sessionBytes)

	sm.lock.Lock()

	newSession := ActiveSession{
		user:      user,
		key:       key,
		not_after: time.Now().UnixMilli() + SESSION_EXPIRATION_TIME,
	}

	sm.sessions[sessionId] = &newSession

	sm.lock.Unlock()

	// Call task manager to start pending tasks
	sm.vault.tasks.OnNewSession(&newSession)

	return sessionId
}

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
