// API Tools

package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	JSON_BODY_MAX_LENGTH     = 5 * 1024 * 1024
	AUTH_API_BODY_MAX_LENGTH = 16 * 1024
)

func GetSessionFromRequest(request *http.Request) *ActiveSession {
	sessionToken := request.Header.Get("x-session-token")

	if sessionToken == "" && request.Method == "GET" {
		c, err := request.Cookie("x-session-token")

		if err == nil || c != nil {
			sessionToken = c.Value
		}
	}

	return GetVault().sessions.FindSession(sessionToken)
}

func ParseRangeHeader(request *http.Request) (int64, int64) {
	rangeHeader := request.Header.Get("Range")

	m, _ := regexp.MatchString("bytes=([0-9]*)-([0-9]*)", rangeHeader)

	if !m {
		return 0, 0
	}

	rangeVal := strings.Split(rangeHeader, "=")[1]

	parts := strings.Split(rangeVal, "-")

	var err error

	start := int64(-1)
	end := int64(-1)

	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)

		if err != nil {
			start = -1
		}
	}

	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)

		if err != nil {
			end = -1
		}
	}

	return start, end
}

type APIErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ReturnAPIError(response http.ResponseWriter, status int, code string, message string) {
	var m APIErrorResponse

	m.Code = code
	m.Message = message

	jsonRes, err := json.Marshal(m)

	if err != nil {
		LogError(err)
		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(status)
	response.Write(jsonRes)
}
