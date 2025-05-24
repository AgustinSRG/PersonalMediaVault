// Disk usage API

package main

import (
	"encoding/json"
	"net/http"

	du "github.com/AgustinSRG/go-disk-usage"
)

type DiskUsage struct {
	Usage     float32 `json:"usage"`
	Available uint64  `json:"available"`
	Free      uint64  `json:"free"`
	Total     uint64  `json:"total"`
}

func api_getDiskUsage(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	result := DiskUsage{
		Usage:     0.0,
		Available: 0,
		Free:      0,
		Total:     0,
	}

	usage := du.NewDiskUsage(GetVault().path)

	if usage != nil {
		result.Usage = usage.Usage() * 100
		result.Available = usage.Available()
		result.Free = usage.Free()
		result.Total = usage.Size()
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
