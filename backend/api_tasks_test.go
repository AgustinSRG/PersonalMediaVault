// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func Tasks_API_Test(server *httptest.Server, session string, t *testing.T) {
	ready := false

	for !ready {
		statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/tasks", nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		taskList := make([]*TaskListInfoEntry, 0)

		err = json.Unmarshal(bodyResponseBytes, &taskList)

		if err != nil {
			t.Error(err)
			return
		}

		ready = len(taskList) == 0
	}
}
