// API Tools

package main

import (
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"os"
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

	if CORS_INSECURE_MODE_ENABLED && sessionToken == "" && (request.Method == "GET" || request.Method == "HEAD") {
		sessionToken = request.URL.Query().Get("x-session-token")
	}

	if sessionToken == "" && (request.Method == "GET" || request.Method == "HEAD") {
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
	response.Header().Add("Cache-Control", "no-cache")
	response.WriteHeader(status)
	response.Write(jsonRes)
}

func GetClientIP(request *http.Request) string {
	ip, _, _ := net.SplitHostPort(request.RemoteAddr)

	if os.Getenv("USING_PROXY") == "YES" {
		forwardedFor := request.Header.Get("X-Forwarded-For")

		if forwardedFor != "" {
			return forwardedFor
		} else {
			return ip
		}
	} else {
		return ip
	}
}

func GetExtensionFromFileName(fileName string) string {
	parts := strings.Split(fileName, ".")

	if len(parts) > 1 {
		ext := strings.ToLower(parts[len(parts)-1])

		r := regexp.MustCompile("[^a-z0-9]+")

		ext = r.ReplaceAllString(ext, "")

		if ext != "" {
			return ext
		} else {
			return "bin"
		}
	} else {
		return "bin"
	}
}

func GetNameFromFileName(fileName string) string {
	parts := strings.Split(fileName, ".")

	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	} else {
		return fileName
	}
}

func EncryptOriginalAssetFile(mid uint64, file string, key []byte) (string, error) {
	encrypted_file := GetTemporalFileName("pma", true)

	f, err := os.OpenFile(file, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return "", err
	}

	f_info, err := f.Stat()

	if err != nil {
		f.Close()

		return "", err
	}

	ws, err := CreateFileBlockEncryptWriteStream(encrypted_file)

	if err != nil {
		f.Close()

		return "", err
	}

	err = ws.Initialize(f_info.Size(), key)

	if err != nil {
		ws.Close()
		f.Close()

		os.Remove(encrypted_file)

		return "", err
	}

	finished := false
	buf := make([]byte, 1024*1024)
	bytesEncrypted := 0

	for !finished {
		c, err := f.Read(buf)

		if err != nil && err != io.EOF {
			ws.Close()
			f.Close()

			os.Remove(encrypted_file)

			GetVault().media.EndProgress(mid)

			return "", err
		}

		if err == io.EOF {
			finished = true
		}

		if c == 0 {
			continue
		}

		err = ws.Write(buf[:c])

		if err != nil {
			LogError(err)

			ws.Close()
			f.Close()

			GetVault().media.EndProgress(mid)

			os.Remove(encrypted_file)

			return "", err
		}

		bytesEncrypted += c

		progress_enc := math.Round(float64(bytesEncrypted) * 100 / float64(ws.file_size))
		p := int32(progress_enc)
		GetVault().media.SetProgress(mid, p)
	}

	ws.Close()
	f.Close()

	GetVault().media.EndProgress(mid)

	return encrypted_file, nil
}

func EncryptAssetFile(file string, key []byte) (string, error) {
	encrypted_file := GetTemporalFileName("pma", true)

	f, err := os.OpenFile(file, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return "", err
	}

	f_info, err := f.Stat()

	if err != nil {
		f.Close()

		return "", err
	}

	ws, err := CreateFileBlockEncryptWriteStream(encrypted_file)

	if err != nil {
		f.Close()

		return "", err
	}

	err = ws.Initialize(f_info.Size(), key)

	if err != nil {
		ws.Close()
		f.Close()

		os.Remove(encrypted_file)

		return "", err
	}

	finished := false
	buf := make([]byte, 1024*1024)

	for !finished {
		c, err := f.Read(buf)

		if err != nil && err != io.EOF {
			ws.Close()
			f.Close()

			os.Remove(encrypted_file)

			return "", err
		}

		if err == io.EOF {
			finished = true
		}

		if c == 0 {
			continue
		}

		err = ws.Write(buf[:c])

		if err != nil {
			LogError(err)

			ws.Close()
			f.Close()

			os.Remove(encrypted_file)

			return "", err
		}
	}

	ws.Close()
	f.Close()

	return encrypted_file, nil
}
