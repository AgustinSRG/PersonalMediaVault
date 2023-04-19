// API Test

package main

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
)

func LoginTest(server *httptest.Server) (session string, fingerprint string, e error) {
	body, err := json.Marshal(LoginAPIBody{
		Username: VAULT_DEFAULT_USER,
		Password: VAULT_DEFAULT_PASSWORD,
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
