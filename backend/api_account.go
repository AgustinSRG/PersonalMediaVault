// Account API

package main

import (
	"encoding/json"
	"net/http"
)

type ChangeUsernameBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordBody struct {
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

func api_changeUsername(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p ChangeUsernameBody

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

	// Check password
	valid := GetVault().credentials.CheckCredentials(session.user, p.Password)

	if !valid {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Change username
	err = GetVault().credentials.ChangeUsername(session.user, p.Username)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	err = GetVault().credentials.SaveCredentials()

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	GetVault().sessions.ChangeUsername(session.user, p.Username)

	response.WriteHeader(200)
	return
}

func api_changePassword(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p ChangePasswordBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.OldPassword) == 0 || len(p.OldPassword) > 255 {
		response.WriteHeader(400)
		return
	}

	if len(p.Password) == 0 || len(p.Password) > 255 {
		response.WriteHeader(400)
		return
	}

	// Check password
	valid := GetVault().credentials.CheckCredentials(session.user, p.OldPassword)

	if !valid {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Change password
	err = GetVault().credentials.SetCredentials(session.user, p.Password, session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	err = GetVault().credentials.SaveCredentials()

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	response.WriteHeader(200)
	return
}
