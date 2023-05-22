// Media assets API

package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func api_handleAssetGet(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	onlyHeader := request.Method == "HEAD"

	// Range header

	start, end := ParseRangeHeader(request)

	// Params

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	asset_id, err := strconv.ParseUint(vars["asset"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	session := GetSessionFromRequestCookie(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	filename := vars["filename"]

	match, _ := regexp.MatchString("[a-z0-9\\_]\\.[a-z0-9]+", filename)

	if !match {
		response.WriteHeader(404)
		return
	}

	ext := strings.Split(filename, ".")[1]

	mimeType := "application/octet-stream"

	switch ext {
	// Image
	case "png":
		mimeType = "image/png"
	case "jpeg":
	case "jpg":
		mimeType = "image/jpeg"
	case "gif":
		mimeType = "image/gif"
	case "bmp":
		mimeType = "image/bmp"
	case "webp":
		mimeType = "image/webp"
	// Audio
	case "aac":
		mimeType = "audio/aac"
	case "mid":
	case "midi":
		mimeType = "audio/midi"
	case "mp3":
		mimeType = "audio/mpeg"
	case "oga":
		mimeType = "audio/ogg"
	case "opus":
		mimeType = "audio/opus"
	case "wav":
		mimeType = "audio/wav"
	case "weba":
		mimeType = "audio/webm"
	// Video
	case "avi":
		mimeType = "video/x-msvideo"
	case "mpeg":
		mimeType = "video/mpeg"
	case "mp4":
		mimeType = "video/mp4"
	case "ogv":
		mimeType = "video/ogg"
	case "ts":
		mimeType = "video/mp2t"
	case "webm":
		mimeType = "video/webm"
	case "mkv":
		mimeType = "video/x-matroska"
		// Subtitles
	case "srt":
		mimeType = "application/x-subrip"
	case "json":
		mimeType = "application/json"
	}

	media := GetVault().media.AcquireMediaResource(media_id)
	found, asset_path, asset_lock := media.AcquireAsset(asset_id, ASSET_SINGLE_FILE)

	if !found {
		GetVault().media.ReleaseMediaResource(media_id)
		response.WriteHeader(404)
		return
	}

	asset_lock.StartRead() // Start reading the asset

	s, err := CreateFileBlockEncryptReadStream(asset_path, session.key)

	if err != nil {
		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	fileSeek := int64(0)
	fileEnding := s.file_size - 1
	contentLength := s.file_size
	hasRange := false

	if start < 0 && end > 0 {
		// Only end
		fileSeek = s.file_size - end
		contentLength = end
		hasRange = true
	} else if start >= 0 && end < 0 {
		// Only start point
		fileSeek = start
		contentLength = s.file_size - start
		hasRange = true
	} else if start >= 0 && end > 0 {
		// Both start and end
		fileSeek = start
		fileEnding = end
		contentLength = end - start + 1
		hasRange = true
	}

	if fileSeek < 0 || fileSeek >= s.file_size || fileEnding >= s.file_size || fileEnding < fileSeek {
		// Invalid range
		s.Close()

		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.Header().Add("Content-Range", "bytes */"+fmt.Sprint(s.file_size))

		response.WriteHeader(416)
		return
	}

	if fileSeek > 0 {
		_, err := s.Seek(fileSeek, 0)
		if err != nil {
			// Seek error
			s.Close()

			asset_lock.EndRead()
			media.ReleaseAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			LogError(err)
			response.WriteHeader(500)
			return
		}
	}

	// Send response

	response.Header().Add("Content-Type", mimeType)
	response.Header().Add("Content-Length", fmt.Sprint(contentLength))
	response.Header().Add("Cache-Control", "max-age=31536000")

	if hasRange {
		response.Header().Add("Content-Range", "bytes "+fmt.Sprint(fileSeek)+"-"+fmt.Sprint(fileEnding)+"/"+fmt.Sprint(s.file_size))
		response.WriteHeader(206)
	} else {
		response.WriteHeader(200)
	}

	if onlyHeader {
		s.Close()

		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		return
	}

	// Setup buffer
	buf := make([]byte, 1024*1024)
	bytesRead := int64(0)

	// Read

	for bytesRead < contentLength {
		c, err := s.Read(buf)

		if err != nil {
			LogError(err)

			s.Close()

			asset_lock.EndRead()
			media.ReleaseAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		if c <= 0 {
			s.Close()

			asset_lock.EndRead()
			media.ReleaseAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		bytesToRead := int64(c)

		if bytesToRead > (contentLength - bytesRead) {
			bytesToRead = contentLength - bytesRead
		}

		bytesRead += bytesToRead

		_, err = response.Write(buf[:bytesToRead])

		if err != nil {
			s.Close()

			asset_lock.EndRead()
			media.ReleaseAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}
	}

	s.Close()

	asset_lock.EndRead()
	media.ReleaseAsset(asset_id)
	GetVault().media.ReleaseMediaResource(media_id)
}

func api_handleAssetVideoPreviews(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	asset_id, err := strconv.ParseUint(vars["asset"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	session := GetSessionFromRequestCookie(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	filename := vars["filename"]

	match, _ := regexp.MatchString("[a-z0-9]+\\_[0-9]+\\.jpg", filename)

	if !match {
		response.WriteHeader(404)
		return
	}

	preview_img_index, err := strconv.ParseInt(strings.Split(strings.Split(filename, ".")[0], "_")[1], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	media := GetVault().media.AcquireMediaResource(media_id)
	found, asset_path, asset_lock := media.AcquireAsset(asset_id, ASSET_MULTI_FILE)

	if !found {
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	asset_lock.StartRead()

	s, err := CreateMultiFilePackReadStream(asset_path)

	if err != nil {
		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	if preview_img_index < 0 || preview_img_index >= s.file_count {
		s.Close()

		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	// Read raw data

	b, err := s.GetFile(preview_img_index)

	if err != nil {
		LogError(err)

		s.Close()

		asset_lock.EndRead()
		media.ReleaseAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	s.Close()

	asset_lock.EndRead()
	media.ReleaseAsset(asset_id)
	GetVault().media.ReleaseMediaResource(media_id)

	// Decrypt the data

	d, err := decryptFileContents(b, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Return the image

	response.Header().Add("Content-Type", "image/jpg")
	response.Header().Add("Content-Length", fmt.Sprint(len(d)))
	response.Header().Add("Cache-Control", "max-age=31536000")

	response.WriteHeader(200)

	response.Write(d) //nolint:errcheck
}
