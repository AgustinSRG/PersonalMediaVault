// Media audios API

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

type AudioTrackAPIResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func api_addMediaAudioTrack(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	audioId := request.URL.Query().Get("id")

	if audioId == "" || len(audioId) > 255 {
		ReturnAPIError(response, 400, "INVALID_ID", "Invalid audio track ID")
		return
	}

	audioName := request.URL.Query().Get("name")

	if audioName == "" || len(audioName) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid audio track name")
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
	finished := false

	for !finished {
		n, err := part.Read(buf)

		if err != nil && err != io.EOF {
			LogError(err)

			_ = f.Close()
			DeleteTemporalFile(tempFile)

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

			_ = f.Close()
			DeleteTemporalFile(tempFile)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	_ = f.Close()

	// Probe uploaded file

	probeRes, err := ProbeMediaFileWithFFProbe(tempFile)

	if err != nil {
		LogError(err)
		DeleteTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_AUDIO", "Invalid audio file provided")
		return
	}

	if probeRes.Type != MediaTypeAudio || !probeRes.Encoded {
		DeleteTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_AUDIO", "Invalid audio file provided")
		return
	}

	// Encrypt the audio file

	audio_encrypted_file, err := EncryptAssetFile(tempFile, session.key)

	DeleteTemporalFile(tempFile)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Put the audio into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		_ = os.Remove(audio_encrypted_file)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		_ = os.Remove(audio_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()

		_ = os.Remove(audio_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	if meta.Type != MediaTypeVideo {
		media.CancelWrite()

		_ = os.Remove(audio_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 400, "NOT_SUPPORTED", "This feature is not supported for the media type. Only for videos.")
		return
	}

	audio_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(audio_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()

		_ = os.Remove(audio_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(audio_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(audio_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()

		_ = os.Remove(audio_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change metadata
	audioIndex := meta.FindAudioTrack(audioId)

	if audioIndex == -1 {
		meta.AddAudioTrack(audioId, audioName, audio_asset)
	} else {
		// Remove old asset
		oldAsset := meta.AudioTracks[audioIndex].Asset
		success, asset_path, asset_lock = media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			_ = os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}

		// Set new values
		meta.AudioTracks[audioIndex].Asset = audio_asset
		meta.AudioTracks[audioIndex].Name = audioName
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

	var result AudioTrackAPIResponse

	result.Id = audioId
	result.Name = audioName
	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(audio_asset) + "/audio.mp3" + "?fp=" + GetVault().credentials.GetFingerprint()

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_removeMediaAudioTrack(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	if !HandleAuthConfirmation(response, request, session, false) {
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	audioId := request.URL.Query().Get("id")

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

	audioIndex := meta.FindAudioTrack(audioId)

	if audioIndex != -1 {
		// Remove old asset
		oldAsset := meta.AudioTracks[audioIndex].Asset
		success, asset_path, asset_lock := media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			_ = os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}
		// Remove entry
		meta.RemoveAudioTrack(audioIndex)
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

type MediaAudioEditBody struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func api_renameMediaAudioTrack(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p MediaAudioEditBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Id == "" || len(p.Id) > 255 {
		ReturnAPIError(response, 400, "INVALID_ID", "Invalid audio track ID")
		return
	}

	if p.Name == "" || len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid audio track name")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	audioId := request.URL.Query().Get("id")

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

	audioIndex := meta.FindAudioTrack(audioId)

	if audioIndex == -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Audio track not found")
		return
	}

	otherAudioIndex := meta.FindAudioTrack(p.Id)

	if otherAudioIndex != audioIndex && otherAudioIndex != -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "INVALID_NAME", "There is another audio with the same name")
		return
	}

	meta.AudioTracks[audioIndex].Id = p.Id
	meta.AudioTracks[audioIndex].Name = p.Name

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
