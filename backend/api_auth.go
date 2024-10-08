// Authentication API

package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type LoginAPIBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Duration string `json:"duration"`
}

type LoginAPIResponse struct {
	SessionId        string `json:"session_id"`
	VaultFingerprint string `json:"vault_fingerprint"`
}

const (
	AUTH_FAIL_COOLDOWN = 5 * 1000
)

var (
	LAST_INVALID_PASSWORD_MAP = make(map[string]int64)
	LAST_INVALID_PASSWORD_MU  = &sync.Mutex{}
)

func api_handleAuthLogin(response http.ResponseWriter, request *http.Request) {
	clientIP := GetClientIP(request)

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p LoginAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Username) == 0 || len(p.Username) > 255 {
		response.WriteHeader(400)
		return
	}

	if len(p.Password) == 0 || len(p.Password) > 255 {
		response.WriteHeader(400)
		return
	}

	// Expiration time
	expirationTime := int64(SESSION_EXPIRATION_TIME_DAY)
	switch p.Duration {
	case "week":
		expirationTime = SESSION_EXPIRATION_TIME_WEEK
	case "month":
		expirationTime = SESSION_EXPIRATION_TIME_MONTH
	case "year":
		expirationTime = SESSION_EXPIRATION_TIME_YEAR
	}

	// Check last failure
	LAST_INVALID_PASSWORD_MU.Lock()
	now := time.Now().UnixMilli()
	lastFailure := LAST_INVALID_PASSWORD_MAP[clientIP]

	if now-lastFailure < AUTH_FAIL_COOLDOWN {
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "COOLDOWN", "You must wait 5 seconds after your last failed login request.")
		return
	}

	// Check credentials
	valid, _ := GetVault().credentials.CheckCredentials(p.Username, p.Password)

	if !valid {
		LAST_INVALID_PASSWORD_MAP[clientIP] = now
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "INVALID_CREDENTIALS", "Invalid credentials provided.")
		LogSecurity("[LOGIN] [From IP: " + GetClientIP(request) + "] Failed login attempt with credentials. Username = " + p.Username)
		return
	}

	LAST_INVALID_PASSWORD_MU.Unlock()

	key, cred_info, err := GetVault().credentials.UnlockVault(p.Username, p.Password)

	if err != nil {
		LogError(err)
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	s, err := GetVault().sessions.CreateSession(p.Username, key, cred_info.root, cred_info.write, expirationTime, "")

	if err != nil {
		LogError(err)
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	var r LoginAPIResponse

	r.SessionId = s
	r.VaultFingerprint = GetVault().credentials.GetFingerprint()

	jsonResult, err := json.Marshal(r)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	LogSecurity("[LOGIN] [From IP: " + GetClientIP(request) + "] Successful login with credentials. Username: " + p.Username)

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_handleAuthLogout(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	GetVault().sessions.CloseSession(session.id)

	response.WriteHeader(200)
}
