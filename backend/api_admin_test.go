// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
)

func Admin_API_Test(server *httptest.Server, session string, t *testing.T) {
	initialPassword := os.Getenv("VAULT_INITIAL_PASSWORD")

	if initialPassword == "" {
		initialPassword = VAULT_DEFAULT_PASSWORD
	}

	// Create account

	body, err := json.Marshal(ApiAdminCreateAccountBody{
		Username: "user2",
		Password: "password2",
		Write:    false,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err := DoTestRequestWithConfirmation(server, "POST", "/api/admin/accounts", body, session, initialPassword, "")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// List accounts

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/admin/accounts", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	response := make([]ApiAdminAccountEntry, 0)

	err = json.Unmarshal(bodyResponseBytes, &response)

	if err != nil {
		t.Error(err)
		return
	}

	foundAccount := false

	for i := 0; i < len(response); i++ {
		if response[i].Username == "user2" {
			if response[i].Write {
				t.Error(ErrorMismatch("Account.Write", fmt.Sprint(response[i].Write), "false"))
			}

			foundAccount = true
			break
		}
	}

	if !foundAccount {
		t.Errorf("The account we just created was not found in the list")
	}

	// Login with new account credentials

	body, err = json.Marshal(LoginAPIBody{
		Username: "user2",
		Password: "password2",
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

	// Delete the account

	body, err = json.Marshal(ApiAdminDeleteAccountBody{
		Username: "user2",
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequestWithConfirmation(server, "POST", "/api/admin/accounts/delete", body, session, initialPassword, "")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// List accounts

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/admin/accounts", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	response = make([]ApiAdminAccountEntry, 0)

	err = json.Unmarshal(bodyResponseBytes, &response)

	if err != nil {
		t.Error(err)
		return
	}

	foundAccount = false

	for i := 0; i < len(response); i++ {
		if response[i].Username == "user2" {
			foundAccount = true
			break
		}
	}

	if foundAccount {
		t.Errorf("The account was not deleted after calling the delete API")
	}
}
