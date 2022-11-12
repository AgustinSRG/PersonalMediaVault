// Tasks API

package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func api_getTasks(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	allTasks := GetVault().tasks.GetAllTasks()

	jsonResult, err := json.Marshal(allTasks)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_getTask(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	vars := mux.Vars(request)

	task_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	taskInfo := GetVault().tasks.GetTaskInfo(task_id)

	if taskInfo == nil {
		ReturnAPIError(response, 404, "TASK_NOT_FOUND", "Task not found.")
		return
	}

	jsonResult, err := json.Marshal(taskInfo)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_killTask(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	if !session.write {
		response.WriteHeader(403)
		return
	}

	vars := mux.Vars(request)

	task_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	GetVault().tasks.KillTask(task_id)

	response.WriteHeader(200)
	return
}
