// Test utils for HTTP server

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func GetTestServer() (*httptest.Server, error) {
	err := InitializeTestVault()
	if err != nil {
		return nil, err
	}
	return httptest.NewServer(apiRouter), nil
}

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

func DoTestRequest(server *httptest.Server, method string, path string, body []byte, session string) (statusCode int, bodyResponse []byte, e error) {
	client := server.Client()

	apiURL, err := url.JoinPath(server.URL, path)

	if err != nil {
		return 0, nil, err
	}

	var bodyReader io.Reader = nil

	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest("POST", apiURL, bodyReader)

	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("x-session-token", session)

	resp, err := client.Do(req)

	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)

	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, bodyData, nil
}
