// Account API

package main

import (
	"encoding/json"
	"net/http"
)

type UsernameInfoAPIResponse struct {
	Username string `json:"username"`
	Root     bool   `json:"root"`
	Write    bool   `json:"write"`
}

type ChangeUsernameBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordBody struct {
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

func api_getUsername(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	var result UsernameInfoAPIResponse

	result.Username = session.user
	result.Root = session.root
	result.Write = session.write

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_changeUsername(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	if !session.root {
		response.WriteHeader(403)
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
		ReturnAPIError(response, 400, "USERNAME_INVALID", "Invalid username.")
		return
	}

	if len(p.Password) == 0 || len(p.Password) > 255 {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Check password
	valid, _ := GetVault().credentials.CheckCredentials(session.user, p.Password)

	if !valid {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Check username
	exists := GetVault().credentials.CheckUserExists(p.Username)

	if exists {
		ReturnAPIError(response, 400, "USERNAME_IN_USE", "The username is already in use.")
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
	valid, cred_info := GetVault().credentials.CheckCredentials(session.user, p.OldPassword)

	if !valid {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Change password
	if cred_info.root && session.root {
		err = GetVault().credentials.SetRootCredentials(session.user, p.Password, session.key)
	} else {
		err = GetVault().credentials.SetAccountCredentials(session.user, p.Password, session.key, session.write)
	}

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
