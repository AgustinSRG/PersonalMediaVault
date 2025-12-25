// Media attachments API

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type AttachmentAPIResponse struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
	Size uint64 `json:"size"`
	Url  string `json:"url"`
}

func api_addMediaAttachment(response http.ResponseWriter, request *http.Request) {
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

	if fileName == "" {
		fileName = "file"
	}

	if len(fileName) > 255 {
		fileName = fileName[:255]
	}

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
	size := uint64(0)

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

		size += uint64(n)

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

	// Encrypt the attachment file

	attachment_encrypted_file, err := EncryptAssetFile(tempFile, session.key)

	DeleteTemporalFile(tempFile)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Put the audio into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		_ = os.Remove(attachment_encrypted_file)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		_ = os.Remove(attachment_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()

		_ = os.Remove(attachment_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	attachment_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(attachment_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()

		_ = os.Remove(attachment_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(attachment_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(attachment_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()

		_ = os.Remove(attachment_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change metadata
	meta.AddAttachment(fileName, attachment_asset, size)

	// Save
	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	var result AttachmentAPIResponse

	result.Id = attachment_asset
	result.Name = fileName
	result.Size = size
	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(attachment_asset) + "/" + url.PathEscape(fileName) + "?fp=" + GetVault().credentials.GetFingerprint()

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type MediaAttachmentEditNameBody struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func api_updateMediaAttachment(response http.ResponseWriter, request *http.Request) {
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

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p MediaAttachmentEditNameBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Name) == 0 || len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid name provided")
		return
	}

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

	attachmentIndex := meta.FindAttachment(p.Id)

	if attachmentIndex == -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Attachment not found")
		return
	}

	meta.Attachments[attachmentIndex].Name = p.Name

	var result AttachmentAPIResponse

	result.Id = meta.Attachments[attachmentIndex].Asset
	result.Name = meta.Attachments[attachmentIndex].Name
	result.Size = meta.Attachments[attachmentIndex].Size
	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.Attachments[attachmentIndex].Asset) + "/" + url.PathEscape(meta.Attachments[attachmentIndex].Name) + "?fp=" + GetVault().credentials.GetFingerprint()

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_removeMediaAttachment(response http.ResponseWriter, request *http.Request) {
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

	attachmentId, err := strconv.ParseUint(request.URL.Query().Get("id"), 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

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

	attachmentIndex := meta.FindAttachment(attachmentId)

	if attachmentIndex != -1 {
		// Remove old asset
		oldAsset := meta.Attachments[attachmentIndex].Asset
		success, asset_path, asset_lock := media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			_ = os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}
		// Remove entry
		meta.RemoveAttachment(attachmentIndex)
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
