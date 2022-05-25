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
		response.WriteHeader(401)
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

		response.WriteHeader(500)
		return
	}

	fileName := part.FileName()
	ext := GetExtensionFromFileName(fileName)

	tempFile := GetTemporalFileName(ext)

	// Write to temp file

	f, err := os.OpenFile(tempFile, os.O_WRONLY, FILE_PERMISSION)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
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

			response.WriteHeader(500)
			return
		}

		if n <= 0 || err == io.EOF {
			finished = true
			continue
		}

		_, err = f.Write(buf[:n])

		if err != nil {
			LogError(err)

			f.Close()
			WipeTemporalFile(tempFile)

			response.WriteHeader(500)
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

		response.WriteHeader(500)
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

		response.WriteHeader(500)
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
	err = os.Rename(thumb_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(thumb_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(500)
		return
	}

	// Change metadata
	meta.ThumbnailReady = true
	meta.ThumbnailAsset = thumb_asset

	// Save
	err = media.EndWrite(meta, session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(500)
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	var result ThumbnailAPIResponse

	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(thumb_asset) + "/thumbnail.jpg"

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.Header().Add("Cache-Control", "no-cache")
	response.WriteHeader(200)

	response.Write(jsonResult)
}
