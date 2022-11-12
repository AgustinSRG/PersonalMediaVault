// Configuration API

package main

import (
	"encoding/json"
	"net/http"
)

func api_getConfig(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	config, err := GetVault().config.Read(session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	jsonResult, err := json.Marshal(config)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_setConfig(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	if !session.root {
		response.WriteHeader(403)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p UserConfig

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Resolutions == nil {
		p.Resolutions = make([]UserConfigResolution, 0)
	}

	err = GetVault().config.Write(&p, session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	err = GetVault().tasks.LoadUserConfigParams(session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	response.WriteHeader(200)
}
