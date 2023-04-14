// API Tools

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
)

const (
	JSON_BODY_MAX_LENGTH     = 5 * 1024 * 1024 // Max length of body for JSON APIs
	AUTH_API_BODY_MAX_LENGTH = 16 * 1024       // Max length of body for authentication requests
	ASSET_JWT_SUB            = "pmv_asset"     // Subject to use for JWT for assets
)

var (
	ASSET_JWT_SECRET = make([]byte, 32) // Secret used to sign tokens for asset requests
)

// Initailizes secret to sign JWT tokens for assets
func InitAssetJWTSecret() {
	rand.Read(ASSET_JWT_SECRET) //nolint:errcheck
}

// Validates asset JWT
// token - JWT to validate
// media_id - Media ID
// asset_id - Asset ID
func CheckAssetToken(token string, media_id uint64, asset_id uint64) (valid bool) {
	if token == "" {
		return false
	}

	defer func() {
		r := recover()
		if r != nil {
			valid = false
		}
	}()

	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Check the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// Provide signing key
		return ASSET_JWT_SECRET, nil
	})

	if err != nil {
		return false
	}

	claims, ok := t.Claims.(jwt.MapClaims)

	if !ok || !t.Valid {
		return false // Invalid token
	}

	if claims["sub"] == nil || claims["sub"].(string) != ASSET_JWT_SUB {
		return false // Invalid sibject
	}

	if claims["mid"] == nil || claims["mid"].(string) != fmt.Sprint(media_id) {
		return false // Invalid media ID
	}

	if claims["aid"] == nil || claims["aid"].(string) != fmt.Sprint(asset_id) {
		return false // Invalid asset ID
	}

	return true
}

// Creates an asset JWT
// media_id - Media ID
// asset_id - Asset ID
func MakeAssetToken(media_id uint64, asset_id uint64) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": ASSET_JWT_SUB,
		"mid": fmt.Sprint(media_id),
		"aid": fmt.Sprint(asset_id),
	})

	tokenb64, e := token.SignedString(ASSET_JWT_SECRET)

	if e != nil {
		return ""
	}

	return tokenb64
}

// Finds session from request headers
// request - HTTP request
// Returns the reference to the session, or nil if unauthorized
func GetSessionFromRequest(request *http.Request) *ActiveSession {
	sessionToken := request.Header.Get("x-session-token")

	return GetVault().sessions.FindSession(sessionToken)
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
	response.Header().Add("Cache-Control", "no-cache")

	if request.Header.Get("If-None-Match") == etag {
		response.WriteHeader(304)
	} else {
		response.Header().Add("Content-Type", "application/json")
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
		response.Header().Add("Cache-Control", "no-cache")
		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.Header().Add("Cache-Control", "no-cache")
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
// Returns a temporal file with the encrypted contents
// This method also sets the progress in the media assets manager
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

func EncryptAssetData(data []byte, key []byte) (string, error) {
	encrypted_file := GetTemporalFileName("pma", true)

	ws, err := CreateFileBlockEncryptWriteStream(encrypted_file)

	if err != nil {

		return "", err
	}

	err = ws.Initialize(int64(len(data)), key)

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
