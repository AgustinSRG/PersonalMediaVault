// Invites API

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type LoginWithInviteCodeBody struct {
	Code string `json:"code"`
}

func api_loginWithInviteCode(response http.ResponseWriter, request *http.Request) {
	clientIP := GetClientIP(request)

	var p LoginWithInviteCodeBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Code == "" {
		ReturnAPIError(response, 400, "INVALID_CODE", "The code must not be blank.")
		return
	}

	if len(p.Code) != 6 {
		ReturnAPIError(response, 400, "INVALID_CODE", "The code must be 6 characters long.")
		return
	}

	LAST_INVALID_PASSWORD_MU.Lock()
	now := time.Now().UnixMilli()
	lastFailure := LAST_INVALID_PASSWORD_MAP[clientIP]

	if now-lastFailure < AUTH_FAIL_COOLDOWN {
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "COOLDOWN", "You must wait 5 seconds after your last failed login request.")
		return
	}

	// Check code
	valid, invitedBy, key, duration := GetVault().invites.UseCode(p.Code)

	if !valid {
		LAST_INVALID_PASSWORD_MAP[clientIP] = now
		LAST_INVALID_PASSWORD_MU.Unlock()
		LogSecurity("[LOGIN] [From IP: " + GetClientIP(request) + "] Failed login attempt with an invite code")
		ReturnAPIError(response, 403, "INVALID_CODE", "Invalid code provided.")
		return
	}

	LAST_INVALID_PASSWORD_MU.Unlock()

	s, err := GetVault().sessions.CreateSession(CreateSessionOptions{
		user:           "",
		key:            key,
		root:           false,
		write:          false,
		expirationTime: duration,
		invitedBy:      invitedBy,
	})

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

	LogSecurity("[LOGIN] [From IP: " + GetClientIP(request) + "] Successful login with invite code. Invited By: " + invitedBy)

	ReturnAPI_JSON(response, request, jsonResult)
}

type InviteStatusAPIResponse struct {
	HasCode             bool   `json:"has_code"`
	Code                string `json:"code"`
	Duration            int64  `json:"duration"`
	ExpirationRemaining int64  `json:"expiration_remaining"`
}

func api_getInviteCodeStatus(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		// Invited users cannot invite
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	has_code, code, not_after, duration := GetVault().invites.GetCodeByUser(session.GetUser())

	now := time.Now().UnixMilli()

	var expiration_remaining int64

	if now > not_after {
		expiration_remaining = 0
	} else {
		expiration_remaining = not_after - now
	}

	result := InviteStatusAPIResponse{
		HasCode:             has_code,
		Code:                code,
		Duration:            duration,
		ExpirationRemaining: expiration_remaining,
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type InviteCodeSessionItem struct {
	Index      uint64 `json:"index"`
	Timestamp  int64  `json:"timestamp"`
	Expiration int64  `json:"expiration"`
}

func api_getInviteCodeSessions(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		// Invited users cannot invite
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	result := GetVault().sessions.FindInviteSessions(session.GetUser())

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type GenerateInviteCodeBody struct {
	Duration string `json:"duration"`
}

const DEFAULT_INVITE_LIMIT = 10

func api_generateInviteCode(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		// Invited users cannot invite
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	config, err := GetVault().config.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	inviteLimit := int(config.InviteLimitPerUser)

	if inviteLimit <= 0 {
		inviteLimit = DEFAULT_INVITE_LIMIT
	}

	currentInvites := len(GetVault().sessions.FindInviteSessions(session.GetUser()))

	if currentInvites >= inviteLimit {
		ReturnAPIError(response, 400, "LIMIT", "You reached the limit of invite codes for your account.")
		return
	}

	var p GenerateInviteCodeBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
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

	code, not_after, err := GetVault().invites.GenerateCode(session.GetUser(), session.key, expirationTime)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	now := time.Now().UnixMilli()

	var expiration_remaining int64

	if now > not_after {
		expiration_remaining = 0
	} else {
		expiration_remaining = not_after - now
	}

	result := InviteStatusAPIResponse{
		HasCode:             true,
		Code:                code,
		Duration:            expirationTime,
		ExpirationRemaining: expiration_remaining,
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_clearInviteCode(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		// Invited users cannot invite
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	GetVault().invites.ClearCode(session.GetUser())

	response.WriteHeader(200)
}

func api_closeInviteSession(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		// Invited users cannot invite
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	vars := mux.Vars(request)

	index, err := strconv.ParseUint(vars["index"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	GetVault().sessions.RemoveInviteSession(session.GetUser(), index)

	response.WriteHeader(200)
}
