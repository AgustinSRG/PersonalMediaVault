// API Test

package main

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
)

func LoginTest(server *httptest.Server) (session string, fingerprint string, e error) {
	initialUser := os.Getenv("VAULT_INITIAL_USER")

	if initialUser == "" {
		initialUser = VAULT_DEFAULT_USER
	}

	initialPassword := os.Getenv("VAULT_INITIAL_PASSWORD")

	if initialPassword == "" {
		initialPassword = VAULT_DEFAULT_PASSWORD
	}

	body, err := json.Marshal(LoginAPIBody{
		Username: initialUser,
		Password: initialPassword,
	})

	if err != nil {
		return "", "", err
	}

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "POST", "/api/auth/login", body, "")

	if err != nil {
		return "", "", err
	}

	if statusCode != 200 {
		return "", "", errors.New("Authentication failed")
	}

	bodyResponse := LoginAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &bodyResponse)

	if err != nil {
		return "", "", err
	}

	return bodyResponse.SessionId, bodyResponse.VaultFingerprint, nil
}

func LogoutTest(server *httptest.Server, session string) (e error) {
	_, _, err := DoTestRequest(server, "POST", "/api/auth/logout", nil, session)

	return err
}
