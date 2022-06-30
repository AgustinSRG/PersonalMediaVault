// Upload media API

package main

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
)

type UploadAPIResponse struct {
	Id uint64 `json:"media_id"`
}

func api_uploadMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	userConfig, err := GetVault().config.Read(session.key)

	if err != nil {
		response.WriteHeader(500)
		return
	}

	givenTitle := request.URL.Query().Get("title")

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
	mediaTitle := GetNameFromFileName(fileName)
	ext := GetExtensionFromFileName(fileName)

	if mediaTitle == "" {
		mediaTitle = "Untitled"
	}

	if givenTitle != "" {
		mediaTitle = givenTitle
	}

	if len(mediaTitle) > 255 {
		mediaTitle = mediaTitle[:255]
	}

	tempFile := GetTemporalFileName(ext, false)

	// Write to temp file

	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE, FILE_PERMISSION)

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

		ReturnAPIError(response, 400, "INVALID_MEDIA", "Invalid media file provided")
		return
	}

	// Create initial media

	media_id, err := GetVault().media.NextMediaId()

	if err != nil {
		LogError(err)

		WipeTemporalFile(tempFile)

		response.WriteHeader(500)
		return
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		WipeTemporalFile(tempFile)

		response.WriteHeader(500)
		return
	}

	err = media.CreateNewMediaAsset(session.key, probe_data.Type, mediaTitle, "", probe_data.Duration, probe_data.Width, probe_data.Height, probe_data.Fps)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		WipeTemporalFile(tempFile)

		response.WriteHeader(500)
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	// Add to main index

	err = AddMediaToMainIndex(media_id)

	if err != nil {
		LogError(err)

		WipeTemporalFile(tempFile)

		response.WriteHeader(500)
		return
	}

	// Background tasks

	go func() {
		BackgroundTaskGenerateThumbnail(session, media_id, tempFile, probe_data)
		BackgroundTaskSaveOriginal(session, media_id, tempFile, ext, probe_data, userConfig)

		WipeTemporalFile(tempFile)
	}()

	// Return the new ID to the client

	var result UploadAPIResponse

	result.Id = media_id

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

func BackgroundTaskGenerateThumbnail(session *ActiveSession, media_id uint64, tempFile string, probe_data *FFprobeMediaResult) {
	thumbnail, err := GenerateThumbnailFromMedia(tempFile, probe_data)

	if err != nil {
		LogError(err)

		return
	}

	// Encrypt the thumbnail

	thumb_encrypted_file, err := EncryptAssetFile(thumbnail, session.key)

	WipeTemporalFile(thumbnail)

	if err != nil {
		LogError(err)

		return
	}

	// Put the thumbnail into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(thumb_encrypted_file)
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		os.Remove(thumb_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	thumb_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(thumb_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(thumb_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

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

		return
	}

	// Change metadata
	meta.ThumbnailReady = true
	meta.ThumbnailAsset = thumb_asset

	// Save
	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	GetVault().media.ReleaseMediaResource(media_id)
}

func BackgroundTaskSaveOriginal(session *ActiveSession, media_id uint64, tempFile string, ext string, probe_data *FFprobeMediaResult, userConfig *UserConfig) {
	// Encrypt the original file

	original_encrypted_file, err := EncryptOriginalAssetFile(media_id, tempFile, session.key)

	if err != nil {
		LogError(err)

		return
	}

	// Put the original into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(original_encrypted_file)
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

	if err != nil {
		LogError(err)

		os.Remove(original_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(original_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	original_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(original_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(original_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = os.Rename(original_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(original_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(original_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	// Original
	meta.OriginalReady = true

	if probe_data.Encoded {
		meta.OriginalEncoded = true
		meta.OriginalExtension = probe_data.EncodedExt
		meta.OriginalAsset = original_asset
		meta.OriginalTask = 0
	} else {
		// Must start a task to encode
		meta.OriginalEncoded = false
		meta.OriginalExtension = ext
		meta.OriginalAsset = original_asset
		meta.OriginalTask = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_ORIGINAL, nil)
	}

	if meta.Type == MediaTypeVideo {
		// Previews
		meta.PreviewsReady = false
		meta.PreviewsAsset = 0
		meta.PreviewsInterval = 0
		meta.PreviewsTask = GetVault().tasks.AddTask(session, media_id, TASK_IMAGE_PREVIEWS, nil)
	}

	// Other resolutions
	if meta.Type == MediaTypeVideo && userConfig.Resolutions != nil {
		for i := 0; i < len(userConfig.Resolutions); i++ {
			if userConfig.Resolutions[i].Fits(meta.Width, meta.Height, probe_data.Fps) {
				// Spawn task

				task_id := GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
					Width:  userConfig.Resolutions[i].Width,
					Height: userConfig.Resolutions[i].Height,
					Fps:    userConfig.Resolutions[i].Fps,
				})

				// Save resolution

				resolution := MediaResolution{
					Width:     userConfig.Resolutions[i].Width,
					Height:    userConfig.Resolutions[i].Height,
					Fps:       userConfig.Resolutions[i].Fps,
					Ready:     false,
					Asset:     0,
					Extension: "",
					TaskId:    task_id,
				}

				meta.Resolutions = append(meta.Resolutions, resolution)
			}
		}
	} else if meta.Type == MediaTypeImage && userConfig.ImageResolutions != nil {
		for i := 0; i < len(userConfig.ImageResolutions); i++ {
			if userConfig.ImageResolutions[i].Fits(meta.Width, meta.Height) {
				// Spawn task

				task_id := GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
					Width:  userConfig.ImageResolutions[i].Width,
					Height: userConfig.ImageResolutions[i].Height,
					Fps:    1,
				})

				// Save resolution

				resolution := MediaResolution{
					Width:     userConfig.Resolutions[i].Width,
					Height:    userConfig.Resolutions[i].Height,
					Fps:       1,
					Ready:     false,
					Asset:     0,
					Extension: "",
					TaskId:    task_id,
				}

				meta.Resolutions = append(meta.Resolutions, resolution)
			}
		}
	}

	// Save
	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	GetVault().media.ReleaseMediaResource(media_id)
}

func AddMediaToMainIndex(media_id uint64) error {
	main_index, err := GetVault().index.StartWrite()

	if err != nil {
		return err
	}

	_, _, err = main_index.file.AddValue(media_id)

	if err != nil {
		GetVault().index.CancelWrite(main_index)

		return err
	}

	err = GetVault().index.EndWrite(main_index)

	return err
}
