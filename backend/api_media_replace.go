// Media thumbnail update (by the user)

package main

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func api_replaceMedia(response http.ResponseWriter, request *http.Request) {
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
	ext := GetExtensionFromFileName(fileName)

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

			f.Close()
			DeleteTemporalFile(tempFile)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	f.Close()

	// Check auth confirmation

	if !HandleAuthConfirmation(response, request, session, false) {
		DeleteTemporalFile(tempFile)
		return
	}

	// Probe uploaded file

	probe_data, err := ProbeMediaFileWithFFProbe(tempFile)

	if err != nil {
		LogError(err)

		DeleteTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_MEDIA", "Invalid media file provided")
		return
	}

	// Encrypt the uploaded file

	media_encrypted_file, err := EncryptAssetFile(tempFile, session.key)

	DeleteTemporalFile(tempFile)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Read metadata

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(media_encrypted_file)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

	if err != nil {
		LogError(err)

		os.Remove(media_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(media_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	if meta.Type != probe_data.Type {
		media.CancelWrite()
		os.Remove(media_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "INVALID_MEDIA_TYPE", "You tried to replace a media with a different type of media.")
		return
	}

	new_original_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(new_original_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(media_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(media_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(new_original_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(media_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Kill tasks
	GetVault().tasks.KillTaskByMedia(media_id)

	// Change metadata
	meta.MediaDuration = probe_data.Duration
	meta.Width = probe_data.Width
	meta.Height = probe_data.Height

	hasAssetToRemove := meta.OriginalReady
	assetToRemove := meta.OriginalAsset

	meta.OriginalReady = true
	meta.OriginalAsset = new_original_asset
	meta.OriginalError = ""

	if probe_data.Encoded {
		meta.OriginalEncoded = true
		meta.OriginalExtension = probe_data.EncodedExt
		meta.OriginalTask = 0
	} else {
		// Must start a task to encode
		meta.OriginalEncoded = false
		meta.OriginalExtension = ext
		meta.OriginalTask = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_ORIGINAL, nil, true)
	}

	// Re-spawn tasks

	// Check previews

	if meta.Type == MediaTypeVideo && !meta.PreviewsReady {
		// Check task

		task_info := GetVault().tasks.GetTaskInfo(meta.PreviewsTask)

		if task_info == nil {
			// Task crashed or was never spawned, make a new one

			meta.PreviewsTask = GetVault().tasks.AddTask(session, media_id, TASK_IMAGE_PREVIEWS, nil, false)
		}
	} else if meta.Type == MediaTypeVideo && meta.PreviewsReady {
		success, asset_path, asset_lock := media.AcquireAsset(meta.PreviewsAsset, ASSET_MULTI_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(meta.PreviewsAsset)
		}

		meta.PreviewsReady = false
		meta.PreviewsTask = GetVault().tasks.AddTask(session, media_id, TASK_IMAGE_PREVIEWS, nil, false)
	}

	// Check resolutions

	if meta.Resolutions != nil {
		for i := 0; i < len(meta.Resolutions); i++ {
			if !meta.Resolutions[i].Ready {
				task_info := GetVault().tasks.GetTaskInfo(meta.Resolutions[i].TaskId)

				if task_info == nil {
					// Task crashed or was never spawned, make a new one

					meta.Resolutions[i].TaskId = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
						Width:  meta.Resolutions[i].Width,
						Height: meta.Resolutions[i].Height,
						Fps:    meta.Resolutions[i].Fps,
					}, false)
				}
			}
		}
	}

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

	// Index (semantic search)

	if meta.OriginalEncoded {
		semanticSearch := GetVault().semanticSearch

		if semanticSearch != nil && semanticSearch.GetStatus().available {
			semanticSearch.RequestMediaIndexing(media_id, session.key, false)
		}
	}

	// Response

	response.WriteHeader(200)
}
