// Logs

package main

import (
	"log"
)

var (
	log_debug_enabled = false
)

func SetDebugLogEnabled(enabled bool) {
	log_debug_enabled = enabled
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

func LogDebug(line string) {
	if log_debug_enabled {
		LogLine("[DEBUG] " + line)
	}
}
