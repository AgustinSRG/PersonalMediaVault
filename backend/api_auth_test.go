// API test

package main

import (
	"encoding/json"
	"testing"
)

func TestAuthAPI(t *testing.T) {
	server, err := GetTestServer()

	if err != nil {
		t.Error(err)
		panic(err)
	}

	defer server.Close()

	/* Login */

	body, err := json.Marshal(LoginAPIBody{
		Username: VAULT_DEFAULT_USER,
		Password: VAULT_DEFAULT_PASSWORD,
	})

	if err != nil {
		t.Error(err)
		panic(err)
	}

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "POST", "/api/auth/login", body, "")

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if statusCode != 200 {
		t.Errorf("Expected status code 200, but found %d", statusCode)
		return
	}

	bodyResponse := LoginAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &bodyResponse)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if bodyResponse.SessionId == "" {
		t.Errorf("Session ID is empty")
		return
	}

	session := bodyResponse.SessionId

	/* Logout */

	statusCode, _, err = DoTestRequest(server, "POST", "/api/auth/logout", nil, session)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if statusCode != 200 {
		t.Errorf("Expected status code 200, but found %d", statusCode)
		return
	}
}
