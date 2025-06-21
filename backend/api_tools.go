// API Tools

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

const (
	JSON_BODY_MAX_LENGTH     = 5 * 1024 * 1024 // Max length of body for JSON APIs
	AUTH_API_BODY_MAX_LENGTH = 16 * 1024       // Max length of body for authentication requests
)

// Finds session from request headers
// request - HTTP request
// Returns the reference to the session, or nil if unauthorized
func GetSessionFromRequest(request *http.Request) *ActiveSession {
	sessionToken := request.Header.Get("x-session-token")

	return GetVault().sessions.FindSession(sessionToken)
}

// Finds session from request headers
// Uses cookie if header not set
// Use only for GET requests, to avoid CSRF
// request - HTTP request
// Returns the reference to the session, or nil if unauthorized
func GetSessionFromRequestCookie(request *http.Request) *ActiveSession {
	sessionToken := request.Header.Get("x-session-token")

	if sessionToken != "" {
		return GetVault().sessions.FindSession(sessionToken)
	}

	sessionToken = request.URL.Query().Get("session_token")

	if sessionToken != "" {
		return GetVault().sessions.FindSession(sessionToken)
	}

	sessionCookieName := "st-" + GetVault().credentials.GetFingerprint()

	cookie, err := request.Cookie(sessionCookieName)

	if err != nil || cookie == nil {
		return nil
	}

	return GetVault().sessions.FindSession(cookie.Value)
}

// Parses range header from request
// request - HTTP request
// Returns:
//
//	1 - Start index
//	2 - Ending index
//
// Note: -1 in the index means not set
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

// Returns API standard JSON response
// response - HTTP response handler
// request - HTTP request handler
// result - JSON result
func ReturnAPI_JSON(response http.ResponseWriter, request *http.Request, result []byte) {
	hasher := sha256.New()
	hasher.Write(result)
	hash := hasher.Sum(nil)
	etag := hex.EncodeToString(hash)

	response.Header().Set("ETag", etag)
	response.Header().Set("Cache-Control", "no-cache")

	if request.Header.Get("If-None-Match") == etag {
		response.WriteHeader(304)
	} else {
		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(200)

		response.Write(result) //nolint:errcheck
	}
}

// API standard error response
type APIErrorResponse struct {
	Code    string `json:"code"`    // Error code
	Message string `json:"message"` // Error message
}

// Returns API standard error message
// response - HTTP response handler
// request - HTTP request handler
// status - HTTP status
// code - Error code
// message - Error message
func ReturnAPIError(response http.ResponseWriter, status int, code string, message string) {
	var m APIErrorResponse

	m.Code = code
	m.Message = message

	jsonRes, err := json.Marshal(m)

	if err != nil {
		LogError(err)
		response.Header().Set("Cache-Control", "no-cache")
		response.WriteHeader(500)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-cache")
	response.WriteHeader(status)
	response.Write(jsonRes) //nolint:errcheck
}

// Gets client IP address
// request - HTTP request
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

// Encrypts original asset file (when uploaded)
// mid - Media ID
// file - Unencrypted file path
// key - Encryption key
// preserve_original - True if the original is being preserved
// second_phase - True if it is the second phase (for original preservation)
// Returns a temporal file with the encrypted contents
// This method also sets the progress in the media assets manager
func EncryptOriginalAssetFile(mid uint64, file string, key []byte, preserve_original bool, second_phase bool) (string, error) {
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

	ws, err := encrypted_storage.CreateFileBlockEncryptWriteStream(encrypted_file, FILE_PERMISSION)

	if err != nil {
		f.Close()

		return "", err
	}

	err = ws.Initialize(f_info.Size(), ENCRYPTED_BLOCK_MAX_SIZE, key)

	if err != nil {
		ws.Close()
		f.Close()

		os.Remove(encrypted_file)

		return "", err
	}

	finished := false
	buf := make([]byte, 1024*1024)
	bytesEncrypted := float64(0)
	bytesTotal := float64(f_info.Size())

	if bytesTotal == 0 {
		bytesTotal = 1
	}

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

		bytesEncrypted += float64(c)

		progress_enc := math.Round(bytesEncrypted * 100 / bytesTotal)
		p := int32(progress_enc)
		if preserve_original {
			p = p / 2

			if second_phase {
				p += 50
			}
		}
		GetVault().media.SetProgress(mid, p)
	}

	ws.Close()
	f.Close()

	if !preserve_original || second_phase {
		GetVault().media.EndProgress(mid)
	}

	return encrypted_file, nil
}

// Encrypts media asset file
// file - File to encrypt
// key - Encryption key
// Returns a temporal file with the encrypted contents
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

	ws, err := encrypted_storage.CreateFileBlockEncryptWriteStream(encrypted_file, FILE_PERMISSION)

	if err != nil {
		f.Close()

		return "", err
	}

	err = ws.Initialize(f_info.Size(), ENCRYPTED_BLOCK_MAX_SIZE, key)

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

func EncryptAssetData(data []byte, key []byte) (string, error) {
	encrypted_file := GetTemporalFileName("pma", true)

	ws, err := encrypted_storage.CreateFileBlockEncryptWriteStream(encrypted_file, FILE_PERMISSION)

	if err != nil {

		return "", err
	}

	err = ws.Initialize(int64(len(data)), ENCRYPTED_BLOCK_MAX_SIZE, key)

	if err != nil {
		ws.Close()

		os.Remove(encrypted_file)

		return "", err
	}

	err = ws.Write(data)

	if err != nil {
		LogError(err)

		ws.Close()

		os.Remove(encrypted_file)

		return "", err
	}

	ws.Close()

	return encrypted_file, nil
}

func LimitStringSize(str string, size int) string {
	if len(str) <= size {
		return str
	} else if size > 0 {
		return str[:size]
	} else {
		return ""
	}
}

// Handles auth confirmation for an API
// response - The response
// request - The request
// session - The session
// onlyTfa - True in case only two factor authentication confirmation is needed
// Returns true only in case of success. It this function returns false, the API handler must return
func HandleAuthConfirmation(response http.ResponseWriter, request *http.Request, session *ActiveSession, onlyTfa bool) bool {
	password := request.Header.Get("x-auth-confirmation-pw")
	tfaCode := request.Header.Get("x-auth-confirmation-tfa")

	success, cooldown, requiredMethod := session.CheckAuthConfirmation(password, tfaCode, onlyTfa)

	if success {
		return true
	}

	if cooldown {
		ReturnAPIError(response, 403, "COOLDOWN", "Auth confirmation is in cooldown")
		return false
	}

	if requiredMethod == "tfa" {
		if len(tfaCode) == 0 {
			ReturnAPIError(response, 403, "AUTH_CONFIRMATION_REQUIRED_TFA", "Auth confirmation is required: Include the two factor authentication code")
		} else {
			ReturnAPIError(response, 403, "INVALID_TFA_CODE", "Invalid two factor authentication code")
		}
	} else {
		if len(password) == 0 {
			ReturnAPIError(response, 403, "AUTH_CONFIRMATION_REQUIRED_PW", "Auth confirmation is required: Include the account password")
		} else {
			ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid account password")
		}
	}

	return false
}
