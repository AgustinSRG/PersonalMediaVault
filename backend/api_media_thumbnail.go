// Media thumbnail update (by the user)

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

type ThumbnailAPIResponse struct {
	Url string `json:"url"`
}

func api_editMediaThumbnail(response http.ResponseWriter, request *http.Request) {
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

	tempFile := GetTemporalFileName("", false)

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
	}

	f.Close()

	// Probe uploaded file

	probe_data, err := ProbeMediaFileWithFFProbe(tempFile)

	if err != nil {
		LogError(err)

		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_THUMBNAIL", "Invalid thumbnail provided")
		return
	}

	// Generate a thumbnail

	thumbnail, err := GenerateThumbnailFromMedia(tempFile, probe_data)

	WipeTemporalFile(tempFile)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 400, "INVALID_THUMBNAIL", "Invalid thumbnail provided")
		return
	}

	if thumbnail == "" {
		ReturnAPIError(response, 400, "INVALID_THUMBNAIL", "Invalid thumbnail provided")
		return
	}

	// Encrypt the thumbnail

	thumb_encrypted_file, err := EncryptAssetFile(thumbnail, session.key)

	WipeTemporalFile(thumbnail)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Put the thumbnail into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(thumb_encrypted_file)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		os.Remove(thumb_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	thumb_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(thumb_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(thumb_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(thumb_asset)

	err2 := GetVault().albums.OnMediaThumbnailUpdate(media_id, session.key)

	if err2 != nil {
		LogError(err)
	}

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change metadata
	hasAssetToRemove := meta.ThumbnailReady
	assetToRemove := meta.ThumbnailAsset
	meta.ThumbnailReady = true
	meta.ThumbnailAsset = thumb_asset

	// Save
	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Remove old asset
	if hasAssetToRemove {
		success, asset_path, asset_lock = media.AcquireAsset(assetToRemove, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(assetToRemove)
		}
	}

	// Release media

	GetVault().media.ReleaseMediaResource(media_id)

	// Clear cache

	GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(media_id)

	// Response

	var result ThumbnailAPIResponse

	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(thumb_asset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
