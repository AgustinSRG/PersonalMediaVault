// Invite codes

package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

// Invite code expiration (10 minutes)
const INVITE_CODE_EXPIRATION = 10 * 60 * 1000

// Invite code
type InviteCode struct {
	code string // Code

	key []byte // Decryption key to create the session

	duration int64 // Duration for the session (Milliseconds)

	not_after int64 // Invite code expiration (Unix milliseconds)
}

// Invitation Manager
type InvitationManager struct {
	vault *Vault // Reference to vault

	lock *sync.Mutex // Mutex to access the data

	codes map[string]*InviteCode // Mapping user -> code
}

// Initialization
func (im *InvitationManager) Initialize(vault *Vault) {
	im.vault = vault
	im.lock = &sync.Mutex{}
	im.codes = make(map[string]*InviteCode)
}

// Uses an invite code
// code - The invite code
// Returns:
//   - success True if successful
//   - invitedBy User who created the code
//   - key Vault decryption key
//   - duration Session duration (milliseconds)
func (im *InvitationManager) UseCode(code string) (success bool, invitedBy string, key []byte, duration int64) {
	im.lock.Lock()
	defer im.lock.Unlock()

	now := time.Now().UnixMilli()

	for u, codeData := range im.codes {
		if codeData.not_after < now {
			continue // Expired code
		}

		if subtle.ConstantTimeCompare([]byte(codeData.code), []byte(code)) == 1 {
			delete(im.codes, u) // Single use
			return true, u, codeData.key, codeData.duration
		}
	}

	return false, "", nil, 0
}

// Gets code by user
// user the user
// Returns:
//   - has_code True if the user has a code
//   - code The code
//   - not_after Code expiration timestamp (Unix milliseconds)
//   - duration Session duration (milliseconds)
func (im *InvitationManager) GetCodeByUser(user string) (has_code bool, code string, not_after int64, duration int64) {
	im.lock.Lock()
	defer im.lock.Unlock()

	c := im.codes[user]

	if c == nil {
		return false, "", 0, 0
	}

	return true, c.code, c.not_after, c.duration
}

// Generates a code for an user
// user - The user
// key - Vault decryption key
// duration - Duration for the invite session
// Returns:
//   - err The error
//   - code The generated invite code
//   - not_after Code expiration (Unix milliseconds)
func (im *InvitationManager) GenerateCode(user string, key []byte, duration int64) (err error, code string, not_after int64) {
	codeBytes := make([]byte, 6)
	_, err_rand := rand.Read(codeBytes)

	if err_rand != nil {
		return err_rand, "", 0
	}

	code = strings.ToUpper(hex.EncodeToString(codeBytes))
	not_after = time.Now().UnixMilli() + INVITE_CODE_EXPIRATION

	im.lock.Lock()
	defer im.lock.Unlock()

	im.codes[user] = &InviteCode{
		code:      code,
		key:       key,
		not_after: not_after,
		duration:  duration,
	}

	return nil, code, not_after
}

// Clears user invite code
// user - The user
func (im *InvitationManager) ClearCode(user string) {
	im.lock.Lock()
	defer im.lock.Unlock()

	if im.codes[user] != nil {
		delete(im.codes, user)
	}
}

// Changes invite code username
// user - Old username
// new_user - New username
func (im *InvitationManager) ChangeUsername(user string, new_user string) {
	im.lock.Lock()
	defer im.lock.Unlock()

	if im.codes[user] != nil {
		im.codes[new_user] = im.codes[user]
		delete(im.codes, user)
	}
}
