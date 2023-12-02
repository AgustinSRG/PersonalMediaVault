// About / version API

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

type AboutResult struct {
	Version       string `json:"version"`
	LastRelease   string `json:"last_release"`
	FFmpegVersion string `json:"ffmpeg_version"`
}

type LastReleaseFile struct {
	Version string `json:"version"`
}

func api_about(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	result := AboutResult{
		Version:       BACKEND_VERSION,
		LastRelease:   BACKEND_VERSION,
		FFmpegVersion: "unknown",
	}

	// Fetch FFmpeg version

	var out_ffmpeg bytes.Buffer
	cmd := exec.Command(FFMPEG_BINARY_PATH, "-version")

	cmd.Stdout = &out_ffmpeg

	err := cmd.Run()

	if err != nil {
		LogError(err)
	} else {
		first_line_parts := strings.Split(strings.Split(out_ffmpeg.String(), "\n")[0], " ")

		if len(first_line_parts) >= 3 {
			result.FFmpegVersion = first_line_parts[2]
		}
	}

	// Fetch latest release

	resp, err := http.Get("https://agustinsrg.github.io/pmv-site/last_release.json")

	if err != nil {
		LogError(err)
	} else {
		var p LastReleaseFile

		err := json.NewDecoder(resp.Body).Decode(&p)

		if err != nil {
			LogError(err)
		} else {
			result.LastRelease = p.Version
		}
	}

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
