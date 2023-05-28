// Test utils for HTTP server

package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

	pathSpl := strings.Split(path, "?")

	actualPath := path
	query := ""

	if len(pathSpl) == 2 {
		actualPath = pathSpl[0]
		query = pathSpl[1]
	}

	apiURL, err := url.JoinPath(server.URL, actualPath)

	if query != "" {
		apiURL = apiURL + "?" + query
	}

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

func DoTestRangeRequest(server *httptest.Server, session string, method string, path string, rangeHeader string) (statusCode int, head http.Header, bodyResponse []byte, e error) {
	client := server.Client()

	pathSpl := strings.Split(path, "?")

	actualPath := path
	query := ""

	if len(pathSpl) == 2 {
		actualPath = pathSpl[0]
		query = pathSpl[1]
	}

	apiURL, err := url.JoinPath(server.URL, actualPath)

	if query != "" {
		apiURL = apiURL + "?" + query
	}

	if err != nil {
		return 0, nil, nil, err
	}

	req, err := http.NewRequest(method, apiURL, nil)

	if err != nil {
		return 0, nil, nil, err
	}

	req.Header.Set("x-session-token", session)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)

	if err != nil {
		return 0, nil, nil, err
	}

	defer resp.Body.Close()

	bodyData, err := io.ReadAll(resp.Body)

	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, bodyData, nil
}
