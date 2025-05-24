// Upload media API

package main

import (
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type UploadAPIResponse struct {
	Id uint64 `json:"media_id"`
}

func api_uploadMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	userConfig, err := GetVault().config.Read(session.key)

	if err != nil {
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
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

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
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
	fileSize := uint64(0)

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

		fileSize += uint64(n)

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

		ReturnAPIError(response, 400, "INVALID_MEDIA", "Invalid media file provided")
		return
	}

	// Create initial media

	media_id, err := GetVault().media.NextMediaId()

	if err != nil {
		LogError(err)

		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	err = media.CreateNewMediaAsset(session.key, probe_data.Type, mediaTitle, "", probe_data.Duration, probe_data.Width, probe_data.Height, probe_data.Fps)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	// Add to main index

	err = AddMediaToMainIndex(media_id)

	if err != nil {
		LogError(err)

		WipeTemporalFile(tempFile)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Add to album

	albumIdQuery := request.URL.Query().Get("album")

	if albumIdQuery != "" {
		album_id, err := strconv.ParseUint(albumIdQuery, 10, 64)

		if err == nil {
			_, err = GetVault().albums.AddMediaToAlbum(album_id, media_id, session.key)

			if err != nil {
				LogError(err)

				WipeTemporalFile(tempFile)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		}
	}

	// Background tasks

	go func() {
		BackgroundTaskGenerateThumbnail(session, media_id, tempFile, probe_data)

		GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(media_id)

		BackgroundTaskSaveOriginal(session, media_id, tempFile, ext, probe_data, fileName, fileSize, userConfig)

		BackgroundTaskExtractSubtitles(session, media_id, tempFile, probe_data)
		BackgroundTaskExtractAudios(session, media_id, tempFile, probe_data)

		WipeTemporalFile(tempFile)
	}()

	// Return the new ID to the client

	var result UploadAPIResponse

	result.Id = media_id

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func BackgroundTaskGenerateThumbnail(session *ActiveSession, media_id uint64, tempFile string, probe_data *FFprobeMediaResult) {
	thumbnail, err := GenerateThumbnailFromMedia(tempFile, probe_data)

	if err != nil {
		LogError(err)

		return
	}

	if thumbnail == "" {
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
	err = RenameAndReplace(thumb_encrypted_file, asset_path)

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

func BackgroundTaskExtractSubtitles(session *ActiveSession, media_id uint64, tempFile string, probe_data *FFprobeMediaResult) {
	if probe_data.Type != MediaTypeVideo && probe_data.Type != MediaTypeAudio {
		return // Not applicable
	}

	tmpPath, files, err := ExtractSubtitlesFiles(tempFile, probe_data)

	if err != nil {
		LogError(err)

		return
	}

	for i := 0; i < len(files); i++ {
		// Encrypt the SRT file

		srt_encrypted_file, err := EncryptAssetFile(files[i].file, session.key)

		if err != nil {
			LogError(err)

			WipeTemporalPath(tmpPath)

			return
		}

		// Put the subtitles into the media assets

		media := GetVault().media.AcquireMediaResource(media_id)

		if media == nil {
			os.Remove(srt_encrypted_file)
			WipeTemporalPath(tmpPath)
			return
		}

		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			os.Remove(srt_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		if meta == nil {
			media.CancelWrite()
			os.Remove(srt_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		srt_asset := meta.NextAssetID
		meta.NextAssetID++

		success, asset_path, asset_lock := media.AcquireAsset(srt_asset, ASSET_SINGLE_FILE)

		if !success {
			media.CancelWrite()
			os.Remove(srt_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

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

			WipeTemporalPath(tmpPath)

			return
		}

		// Change metadata
		if meta.FindSubtitle(files[i].Id) == -1 {
			meta.AddSubtitle(files[i].Id, files[i].Name, srt_asset)
		}

		// Save
		err = media.EndWrite(meta, session.key, false)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		GetVault().media.ReleaseMediaResource(media_id)
	}

	WipeTemporalPath(tmpPath) // Remove temp path for subtitles
}

func BackgroundTaskExtractAudios(session *ActiveSession, media_id uint64, tempFile string, probe_data *FFprobeMediaResult) {
	if probe_data.Type != MediaTypeVideo {
		return // Not applicable
	}

	tmpPath, files, err := ExtractAudioTracks(tempFile, probe_data)

	if err != nil {
		LogError(err)

		return
	}

	for i := 0; i < len(files); i++ {
		// Encrypt the MP3 file

		mp3_encrypted_file, err := EncryptAssetFile(files[i].file, session.key)

		if err != nil {
			LogError(err)

			WipeTemporalPath(tmpPath)

			return
		}

		// Put the subtitles into the media assets

		media := GetVault().media.AcquireMediaResource(media_id)

		if media == nil {
			os.Remove(mp3_encrypted_file)
			WipeTemporalPath(tmpPath)
			return
		}

		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			os.Remove(mp3_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		if meta == nil {
			media.CancelWrite()
			os.Remove(mp3_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		audio_asset := meta.NextAssetID
		meta.NextAssetID++

		success, asset_path, asset_lock := media.AcquireAsset(audio_asset, ASSET_SINGLE_FILE)

		if !success {
			media.CancelWrite()
			os.Remove(mp3_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		asset_lock.RequestWrite()
		asset_lock.StartWrite()

		// Move temp file
		err = RenameAndReplace(mp3_encrypted_file, asset_path)

		asset_lock.EndWrite()

		media.ReleaseAsset(audio_asset)

		if err != nil {
			LogError(err)

			media.CancelWrite()
			os.Remove(mp3_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		// Change metadata
		if meta.FindAudioTrack(files[i].Id) == -1 {
			meta.AddAudioTrack(files[i].Id, files[i].Name, audio_asset)
		}

		// Save
		err = media.EndWrite(meta, session.key, false)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(media_id)

			WipeTemporalPath(tmpPath)

			return
		}

		GetVault().media.ReleaseMediaResource(media_id)
	}

	WipeTemporalPath(tmpPath) // Remove temp path for subtitles
}

func BackgroundTaskSaveOriginal(session *ActiveSession, media_id uint64, tempFile string, ext string, probe_data *FFprobeMediaResult, originalFileName string, originalFileSize uint64, userConfig *UserConfig) {
	// Encrypt the original file

	preservingOriginal := !probe_data.Encoded && userConfig.PreserveOriginalBeforeEncoding

	original_encrypted_file, err := EncryptOriginalAssetFile(media_id, tempFile, session.key, preservingOriginal, false)

	if err != nil {
		LogError(err)
		return
	}

	var original_preserved_encrypted_file string

	if preservingOriginal {
		// Media will be encoded, but user config
		// requires the original to be preserved

		original_preserved_encrypted_file, err = EncryptOriginalAssetFile(media_id, tempFile, session.key, true, true)

		if err != nil {
			LogError(err)
			os.Remove(original_encrypted_file)
			return
		}
	}

	// Put the original into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(original_encrypted_file)
		if original_preserved_encrypted_file != "" {
			os.Remove(original_preserved_encrypted_file)
		}
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

	if err != nil {
		LogError(err)

		os.Remove(original_encrypted_file)

		if original_preserved_encrypted_file != "" {
			os.Remove(original_preserved_encrypted_file)
		}

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	if meta == nil {
		media.CancelWrite()

		os.Remove(original_encrypted_file)

		if original_preserved_encrypted_file != "" {
			os.Remove(original_preserved_encrypted_file)
		}

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	original_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(original_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()

		os.Remove(original_encrypted_file)

		if original_preserved_encrypted_file != "" {
			os.Remove(original_preserved_encrypted_file)
		}

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(original_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(original_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()

		os.Remove(original_encrypted_file)

		if original_preserved_encrypted_file != "" {
			os.Remove(original_preserved_encrypted_file)
		}

		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	// Original
	meta.OriginalReady = true

	// Add attachment if original must be preserved

	if preservingOriginal {
		attachment_asset := meta.NextAssetID
		meta.NextAssetID++

		success, asset_path, asset_lock := media.AcquireAsset(attachment_asset, ASSET_SINGLE_FILE)

		if !success {
			LogError(err)

			media.CancelWrite()

			os.Remove(original_preserved_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		asset_lock.RequestWrite()
		asset_lock.StartWrite()

		// Move temp file
		err = RenameAndReplace(original_preserved_encrypted_file, asset_path)

		asset_lock.EndWrite()

		media.ReleaseAsset(attachment_asset)

		if err != nil {
			LogError(err)

			media.CancelWrite()

			os.Remove(original_preserved_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		meta.AddAttachment(originalFileName, attachment_asset, originalFileSize)
	}

	// Mark as encoded or create task

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
		meta.OriginalTask = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_ORIGINAL, nil, true)
	}

	if meta.Type == MediaTypeVideo {
		// Previews
		meta.PreviewsReady = false
		meta.PreviewsAsset = 0
		meta.PreviewsInterval = 0
		meta.PreviewsTask = GetVault().tasks.AddTask(session, media_id, TASK_IMAGE_PREVIEWS, nil, false)
	}

	// Other resolutions
	if meta.Type == MediaTypeVideo && userConfig.Resolutions != nil {
		for i := 0; i < len(userConfig.Resolutions); i++ {
			adaptedResolution := userConfig.Resolutions[i].Adapt(meta.Width, meta.Height)

			if adaptedResolution.Fits(meta.Width, meta.Height, probe_data.Fps) {
				// Spawn task

				task_id := GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
					Width:  adaptedResolution.Width,
					Height: adaptedResolution.Height,
					Fps:    adaptedResolution.Fps,
				}, false)

				// Save resolution

				resolution := MediaResolution{
					Width:     adaptedResolution.Width,
					Height:    adaptedResolution.Height,
					Fps:       adaptedResolution.Fps,
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
			adaptedResolution := userConfig.ImageResolutions[i].Adapt(meta.Width, meta.Height)

			if adaptedResolution.Fits(meta.Width, meta.Height) {
				// Spawn task

				task_id := GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
					Width:  adaptedResolution.Width,
					Height: adaptedResolution.Height,
					Fps:    1,
				}, false)

				// Save resolution

				resolution := MediaResolution{
					Width:     adaptedResolution.Width,
					Height:    adaptedResolution.Height,
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
