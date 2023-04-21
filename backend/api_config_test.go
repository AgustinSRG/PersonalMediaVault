// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func Config_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Read config

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/config", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res := UserConfig{}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	// Update config

	res.MaxTasks = 2

	body, err := json.Marshal(res)

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/config", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// Check config changes

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/config", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res2 := UserConfig{}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	if res2.MaxTasks != 2 {
		t.Error(ErrorMismatch("MaxTasks", fmt.Sprint(res2.MaxTasks), fmt.Sprint(2)))
	}
}
