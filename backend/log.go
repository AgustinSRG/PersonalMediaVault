// Logs

package main

import (
	"fmt"
	"log"
	"net/http"
)

var (
	log_debug_enabled    = false
	log_requests_enabled = false
)

func SetDebugLogEnabled(enabled bool) {
	log_debug_enabled = enabled
}

func SetRequestLogEnabled(enabled bool) {
	log_requests_enabled = enabled
}

func LogLine(line string) {
	log.Println(line)
}

func LogWarning(line string) {
	LogLine("[WARNING] " + line)
}

func LogInfo(line string) {
	LogLine("[INFO] " + line)
}

func LogError(err error) {
	LogLine("[ERROR] " + err.Error())
}

func LogTaskError(task_id uint64, err string) {
	LogLine("[TASK #" + fmt.Sprint(task_id) + "] [ERROR] " + err)
}

func LogTaskDebug(task_id uint64, err string) {
	if log_debug_enabled {
		LogLine("[TASK #" + fmt.Sprint(task_id) + "] [DEBUG] " + err)
	}
}

func LogDebug(line string) {
	if log_debug_enabled {
		LogLine("[DEBUG] " + line)
	}
}

func LogRequest(r *http.Request) {
	if log_requests_enabled {
		LogLine("[REQUEST] (From: " + GetClientIP(r) + ") " + r.Method + " " + r.URL.Path)
	}
}
