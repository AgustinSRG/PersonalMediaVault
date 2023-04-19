// Test utils for HTTP server

package main

import (
	"bytes"
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

	req, err := http.NewRequest(method, apiURL, bodyReader)

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
