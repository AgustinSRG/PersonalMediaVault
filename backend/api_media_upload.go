// Upload media API

package main

import "net/http"

func api_uploadMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}
