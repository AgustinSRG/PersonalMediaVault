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
}

type LoginAPIResponse struct {
	SessionId string `json:"session_id"`
}

const (
	AUTH_FAIL_COOLDOWN = 5 * 1000
)

var (
	LAST_INVALID_PASSWORD_MAP = make(map[string]int64)
	LAST_INVALID_PASSWORD_MU  = &sync.Mutex{}
)

func api_handleAuthLogin(response http.ResponseWriter, request *http.Request) {
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

	// Check last failure
	LAST_INVALID_PASSWORD_MU.Lock()
	now := time.Now().UnixMilli()
	lastFailure := LAST_INVALID_PASSWORD_MAP[request.RemoteAddr]

	if now-lastFailure < AUTH_FAIL_COOLDOWN {
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "COOLDOWN", "You must wait 5 seconds after your last failed login request.")
		return
	}

	// Check credentials
	valid := GetVault().credentials.CheckCredentials(p.Username, p.Password)

	if !valid {
		LAST_INVALID_PASSWORD_MAP[request.RemoteAddr] = now
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "INVALID_CREDENTIALS", "Invalid credentials provided.")
		return
	}

	LAST_INVALID_PASSWORD_MU.Unlock()

	key, err := GetVault().credentials.UnlockVault(p.Username, p.Password)

	if err != nil {
		LogError(err)
		response.WriteHeader(500)
		return
	}

	s := GetVault().sessions.CreateSession(p.Username, key)

	var r LoginAPIResponse

	r.SessionId = s

	jsonResult, err := json.Marshal(r)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(200)

	response.Write(jsonResult)
}

func api_handleAuthLogout(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	GetVault().sessions.CloseSession(session.id)

	response.WriteHeader(200)
}
