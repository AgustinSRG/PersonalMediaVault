// Media subtitles API

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type SubtitlesAPIResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

const MAX_SUBTITLES_FILE_SIZE = 10 * 1024 * 1024

func api_addMediaSubtitles(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	subtitlesId := request.URL.Query().Get("id")

	if subtitlesId == "" || len(subtitlesId) > 255 {
		ReturnAPIError(response, 400, "INVALID_ID", "Invalid subtitles ID")
		return
	}

	subtitlesName := request.URL.Query().Get("name")

	if subtitlesName == "" || len(subtitlesName) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid subtitles name")
		return
	}

	_, p, err := mime.ParseMediaType(request.Header.Get("Content-Type"))

	if err != nil {
		response.WriteHeader(400)
		return
	}

	boundary := p["boundary"]

	reader := multipart.NewReader(request.Body, boundary)

	part, err := reader.NextPart()

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	fileName := part.FileName()
	ext := GetExtensionFromFileName(fileName)

	tempFile := GetTemporalFileName(ext, false)

	// Write to temp file

	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE, FILE_PERMISSION)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	buf := make([]byte, 1024*1024)
	remaining := int64(MAX_SUBTITLES_FILE_SIZE)
	finished := false

	for !finished {
		n, err := part.Read(buf)

		if err != nil && err != io.EOF {
			LogError(err)

			f.Close()
			WipeTemporalFile(tempFile)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if err == io.EOF {
			finished = true
		}

		if n == 0 {
			continue
		}

		_, err = f.Write(buf[:n])

		if err != nil {
			LogError(err)

			f.Close()
			WipeTemporalFile(tempFile)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		remaining -= int64(n)
		if remaining < 0 {
			f.Close()
			WipeTemporalFile(tempFile)

			response.WriteHeader(413) // Payload too large
			return
		}
	}

	f.Close()

	// Probe uploaded file

	valid := ValidateSubtitlesFile(tempFile)

	if !valid {
		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_SRT", "Invalid srt file provided")
		return
	}

	// Encrypt the srt file

	srt_encrypted_file, err := EncryptAssetFile(tempFile, session.key)

	WipeTemporalFile(tempFile)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Put the srt into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(srt_encrypted_file)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		os.Remove(srt_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(srt_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	if meta.Type != MediaTypeVideo && meta.Type != MediaTypeAudio {
		media.CancelWrite()
		os.Remove(srt_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "NOT_SUPPORTED", "This feature is not supported for the media type. Only for videos and audios.")
		return
	}

	srt_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(srt_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(srt_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(srt_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(srt_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(srt_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change metadata
	subtitlesIndex := meta.FindSubtitle(subtitlesId)

	if subtitlesIndex == -1 {
		meta.AddSubtitle(subtitlesId, subtitlesName, srt_asset)
	} else {
		// Remove old assset
		oldAsset := meta.Subtitles[subtitlesIndex].Asset
		success, asset_path, asset_lock = media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}

		// Set new values
		meta.Subtitles[subtitlesIndex].Asset = srt_asset
		meta.Subtitles[subtitlesIndex].Name = subtitlesName
	}

	// Save
	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	var result SubtitlesAPIResponse

	result.Id = subtitlesId
	result.Name = subtitlesName
	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(srt_asset) + "/subrip.srt" + "?fp=" + GetVault().credentials.GetFingerprint()

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_removeMediaSubtitles(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	subtitlesId := request.URL.Query().Get("id")

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	subtitlesIndex := meta.FindSubtitle(subtitlesId)

	if subtitlesIndex != -1 {
		// Remove old assset
		oldAsset := meta.Subtitles[subtitlesIndex].Asset
		success, asset_path, asset_lock := media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}
		// Remove entry
		meta.RemoveSubtitle(subtitlesIndex)
	}

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

type MediaSubtitlesEditBody struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func api_renameMediaSubtitles(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p MediaSubtitlesEditBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Id == "" || len(p.Id) > 255 {
		ReturnAPIError(response, 400, "INVALID_ID", "Invalid subtitles ID")
		return
	}

	if p.Name == "" || len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid subtitles name")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	subtitlesId := request.URL.Query().Get("id")

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	subtitlesIndex := meta.FindSubtitle(subtitlesId)

	if subtitlesIndex == -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Audio track not found")
		return
	}

	otherSubtitlesIndex := meta.FindSubtitle(p.Id)

	if otherSubtitlesIndex != subtitlesIndex && otherSubtitlesIndex != -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "INVALID_NAME", "There is another subtitles with the same name")
		return
	}

	meta.Subtitles[subtitlesIndex].Id = p.Id
	meta.Subtitles[subtitlesIndex].Name = p.Name

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}
