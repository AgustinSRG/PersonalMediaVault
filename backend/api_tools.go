// API Tools

package main

import "net/http"

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
