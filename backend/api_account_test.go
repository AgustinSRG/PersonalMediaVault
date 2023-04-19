// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func Account_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Check get username API
	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/account/username", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res1 := UsernameInfoAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res1)

	if err != nil {
		t.Error(err)
		return
	}

	if !res1.Root {
		t.Error(ErrorMismatch("Root", fmt.Sprint(res1.Root), "true"))
	}

	if res1.Username != VAULT_DEFAULT_USER {
		t.Error(ErrorMismatch("Username", fmt.Sprint(res1.Username), VAULT_DEFAULT_USER))
	}

	if !res1.Write {
		t.Error(ErrorMismatch("Write", fmt.Sprint(res1.Write), "true"))
	}

	// Check invalid session

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/account/username", nil, "random_invalid_session")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 401 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "401"))
	}

	// Change username

	body, err := json.Marshal(ChangeUsernameBody{
		Username: "test",
		Password: VAULT_DEFAULT_PASSWORD,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "POST", "/api/account/username", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// Change password

	body, err = json.Marshal(ChangePasswordBody{
		Password:    "test_password",
		OldPassword: VAULT_DEFAULT_PASSWORD,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "POST", "/api/account/password", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// Test new credentials

	body, err = json.Marshal(LoginAPIBody{
		Username: "test",
		Password: "test_password",
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "POST", "/api/auth/login", body, "")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	bodyResponse := LoginAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &bodyResponse)

	if err != nil {
		t.Error(err)
		return
	}

	// Close the session

	statusCode, _, err = DoTestRequest(server, "POST", "/api/auth/logout", nil, bodyResponse.SessionId)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}
}
