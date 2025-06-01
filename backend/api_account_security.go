// Account security API

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"
	"time"

	"github.com/pquerna/otp"
)

type AccountSecurityOptions struct {
	TwoFactorAuthEnabled bool   `json:"tfa"`
	TwoFactorAuthMethod  string `json:"tfaMethod"`

	AuthConfirmationEnabled       bool   `json:"authConfirmation"`
	AuthConfirmationMethod        string `json:"authConfirmationMethod"`
	AuthConfirmationPeriodSeconds uint32 `json:"authConfirmationPeriodSeconds"`
}

func api_getSecurityOptions(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	securityOptions := session.GetAccountSecurityOptions()

	jsonResult, err := json.Marshal(securityOptions)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type SetAccountSecurityOptionsBody struct {
	AuthConfirmationEnabled       bool   `json:"authConfirmation"`
	AuthConfirmationMethod        string `json:"authConfirmationMethod"`
	AuthConfirmationPeriodSeconds uint32 `json:"authConfirmationPeriodSeconds"`
}

func api_setSecurityOptions(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p SetAccountSecurityOptionsBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.AuthConfirmationMethod != "pw" {
		p.AuthConfirmationMethod = "tfa"
	}

	user := session.GetUser()

	GetVault().credentials.ChangeSecuritySettings(user, p.AuthConfirmationEnabled, p.AuthConfirmationMethod, p.AuthConfirmationPeriodSeconds)
	GetVault().sessions.UpdateUserSessionsAuthConfirmation(user, p.AuthConfirmationEnabled, p.AuthConfirmationMethod, p.AuthConfirmationPeriodSeconds)

	response.WriteHeader(200)
}

type TimeOtpParametersResult struct {
	Secret string `json:"secret"`
	Method string `json:"method"`
	Url    string `json:"url"`
	Image  string `json:"qr"`
}

const TFA_QR_SIZE_PIXELS = 256

func api_getParametersTwoFactorAuthTimeOtp(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	issuer := request.URL.Query().Get("issuer")

	if issuer == "" {
		issuer = "PersonalMediaVault"
	}

	account := request.URL.Query().Get("account")

	if account == "" {
		account = session.GetUser()
	}

	algoStr := request.URL.Query().Get("algorithm")

	var algorithm otp.Algorithm

	switch algoStr {
	case "sha1", "":
		algorithm = otp.AlgorithmSHA1
	case "sha256":
		algorithm = otp.AlgorithmSHA256
	case "sha512":
		algorithm = otp.AlgorithmSHA512
	default:
		ReturnAPIError(response, 400, "INVALID_ALGORITHM", "Invalid algorithm. Supported ones are: sha1, sha256, sha512")
		return
	}

	allowSkew := request.URL.Query().Get("skew") == "allow"

	periodStr := request.URL.Query().Get("period")

	var period uint

	switch periodStr {
	case "30":
		period = 30
	case "60", "":
		period = 60
	case "120":
		period = 30
	default:
		ReturnAPIError(response, 400, "INVALID_PERIOD", "Invalid period. Supported ones are: 30, 60, 120")
		return
	}

	totpOptions := TimeOtpOptions{
		Algorithm:      algorithm,
		Period:         period,
		AllowClockSkew: allowSkew,
	}

	key, err := generateTimeOtpKey(issuer, account, &totpOptions)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	img, err := key.Image(TFA_QR_SIZE_PIXELS, TFA_QR_SIZE_PIXELS)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	imageBuf := new(bytes.Buffer)

	err = png.Encode(imageBuf, img)
	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	base64Img := base64.StdEncoding.EncodeToString(imageBuf.Bytes())

	result := TimeOtpParametersResult{
		Secret: key.Secret(),
		Method: totpOptions.ToMethodString(),
		Url:    key.URL(),
		Image:  "data:image/png;base64," + base64Img,
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type EnableTimeOtpBody struct {
	Secret            string `json:"secret"`
	Method            string `json:"method"`
	Password          string `json:"password"`
	TwoFactorAuthCode string `json:"code"`
}

func api_enableTwoFactorAuthTimeOtp(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p EnableTimeOtpBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	tfaEnabled, _, _ := session.GetTwoFactorAuth()

	if tfaEnabled {
		ReturnAPIError(response, 400, "TFA_ENABLED", "Two factor authentication is already enabled for your account")
		return
	}

	// Validate account password

	valid, _ := GetVault().credentials.CheckPassword(session.GetUser(), p.Password)

	if !valid {
		ReturnAPIError(response, 403, "INVALID_PASSWORD", "Invalid password.")
		return
	}

	// Validate TFA options

	tfaMethod := p.Method
	tfaKey := []byte(p.Secret)

	if len(tfaKey) == 0 {
		ReturnAPIError(response, 400, "INVALID_TOTP_SECRET", "Invalid TOTP secret provided")
		return
	}

	_, err = parseTimeOtpOptions(p.Method)

	if err != nil {
		ReturnAPIError(response, 400, "INVALID_TOTP_OPTIONS", "Invalid TOTP options provided in the method field")
		return
	}

	valid = validateTwoFactorAuthCode(tfaMethod, tfaKey, p.TwoFactorAuthCode)

	if !valid {
		ReturnAPIError(response, 400, "INVALID_TOTP_CODE", "Invalid TOTP code provided")
		return
	}

	user := session.GetUser()

	GetVault().credentials.EnableTfa(user, tfaMethod, tfaKey, p.Password)
	GetVault().sessions.UpdateUserSessionsEnableTfa(user, tfaKey, tfaMethod)

	response.WriteHeader(200)
}

type DisableTwoFactorAuthBody struct {
	TwoFactorAuthCode string `json:"code"`
}

func api_disableTwoFactorAuth(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.IsUser() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, AUTH_API_BODY_MAX_LENGTH)

	var p DisableTwoFactorAuthBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	tfaEnabled, tfaMethod, tfaKey := session.GetTwoFactorAuth()

	if !tfaEnabled {
		ReturnAPIError(response, 400, "TFA_NOT_ENABLED", "Two factor authentication is not currently enabled for your account")
		return
	}

	clientIP := GetClientIP(request)

	LAST_INVALID_PASSWORD_MU.Lock()
	now := time.Now().UnixMilli()
	lastFailure := LAST_INVALID_PASSWORD_MAP[clientIP]

	if now-lastFailure < AUTH_FAIL_COOLDOWN {
		LAST_INVALID_PASSWORD_MU.Unlock()
		ReturnAPIError(response, 403, "COOLDOWN", "You must wait 5 seconds after your last failed login request.")
		return
	}

	valid := validateTwoFactorAuthCode(tfaMethod, tfaKey, p.TwoFactorAuthCode)

	if !valid {
		LAST_INVALID_PASSWORD_MAP[clientIP] = now
		LAST_INVALID_PASSWORD_MU.Unlock()
		LogSecurity("[TFA Disable failed attempt] [From IP: " + GetClientIP(request) + "] Failed to authorize two factor authentication disable request.")
		ReturnAPIError(response, 403, "INVALID_CODE", "Invalid two factor authentication code.")
		return
	}

	LAST_INVALID_PASSWORD_MU.Unlock()

	user := session.GetUser()

	GetVault().credentials.DisableTfa(user)
	GetVault().sessions.UpdateUserSessionsDisableTfa(user)

	response.WriteHeader(200)
}
