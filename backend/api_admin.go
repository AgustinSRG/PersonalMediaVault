// Accounts admin API

package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type ApiAdminAccountEntry struct {
	Username string `json:"username"`
	Write    bool   `json:"write"`
}

func api_getAccounts(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.root {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	accounts := GetVault().credentials.GetAccountsInfo()

	jsonResult, err := json.Marshal(accounts)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type ApiAdminCreateAccountBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Write    bool   `json:"write"`
}

func api_createAccount(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.root {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p ApiAdminCreateAccountBody

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
		ReturnAPIError(response, 400, "PASSWORD_INVALID", "Invalid password.")
		return
	}

	// Check username
	exists := GetVault().credentials.CheckUserExists(p.Username)

	if exists {
		ReturnAPIError(response, 400, "USERNAME_IN_USE", "The username is already in use.")
		return
	}

	err = GetVault().credentials.SetAccountCredentials(p.Username, p.Password, session.key, p.Write)

	if err != nil {
		LogError(err)
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	err = GetVault().credentials.SaveCredentials()

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	response.WriteHeader(200)
	return
}

type ApiAdminDeleteAccountBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Write    bool   `json:"write"`
}

func api_deleteAccount(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.root {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p ApiAdminDeleteAccountBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Check username
	exists := GetVault().credentials.CheckUserExists(p.Username)

	if !exists {
		ReturnAPIError(response, 404, "USER_NOT_FOUND", "User not found")
		return
	}

	err = GetVault().credentials.RemoveAccount(p.Username)

	if err != nil {
		LogError(err)
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	err = GetVault().credentials.SaveCredentials()

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().sessions.RemoveUserSessions(p.Username)

	response.WriteHeader(200)
	return
}

var (
	LAUNCHER_TAG = ""
)

func api_checkLauncherTag(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	if vars["tag"] == LAUNCHER_TAG {
		response.WriteHeader(200)
		return
	}

	response.WriteHeader(404)
}
