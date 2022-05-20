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
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

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
	}

	media := GetVault().media.AcquireMediaResource(media_id)
	asset_path := media.StartReadAsset(asset_id, ASSET_MUTI_FILE)

	s, err := CreateFileBlockEncryptReadStream(asset_path, session.key)

	if err != nil {
		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	if start >= s.file_size || end >= s.file_size || end <= start {
		// Invalid range
		s.Close()

		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.Header().Add("Content-Range", "bytes */"+fmt.Sprint(s.file_size))

		response.WriteHeader(416)
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

	if fileSeek > 0 {
		s.Seek(fileSeek, 0)
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

		media.EndReadAsset(asset_id)
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

			media.EndReadAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		if c <= 0 {
			s.Close()

			media.EndReadAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}

		bytesRead += int64(c)

		_, err = response.Write(buf[:c])

		if err != nil {
			s.Close()

			media.EndReadAsset(asset_id)
			GetVault().media.ReleaseMediaResource(media_id)

			return
		}
	}

	s.Close()

	media.EndReadAsset(asset_id)
	GetVault().media.ReleaseMediaResource(media_id)
}

func api_handleAssetVideoHLS(response http.ResponseWriter, request *http.Request) {
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

	asset_id, err := strconv.ParseUint(vars["asset"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	filename := vars["filename"]

	match, _ := regexp.MatchString("([a-z0-9]+\\_[0-9]+\\.ts)|([a-z0-9]+\\.m3u8)", filename)

	if !match {
		response.WriteHeader(404)
		return
	}

	parts := strings.Split(filename, ".")
	ext := parts[1]

	file_index := int64(0)

	if ext == "ts" {
		frag_index, err := strconv.ParseInt(strings.Split(parts[0], "_")[1], 10, 64)

		if err != nil || frag_index < 0 {
			response.WriteHeader(400)
			return
		}

		file_index = frag_index + 1
	}

	media := GetVault().media.AcquireMediaResource(media_id)
	asset_path := media.StartReadAsset(asset_id, ASSET_MUTI_FILE)

	s, err := CreateMultiFilePackReadStream(asset_path)

	if err != nil {
		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	if file_index < 0 || file_index >= s.file_count {
		s.Close()

		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	// Read raw data

	b, err := s.GetFile(file_index)

	if err != nil {
		LogError(err)

		s.Close()

		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(500)
		return
	}

	s.Close()

	media.EndReadAsset(asset_id)
	GetVault().media.ReleaseMediaResource(media_id)

	// Decrypt the data

	d, err := decryptFileContents(b, session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	// Return the file

	if ext == "ts" {
		response.Header().Add("Content-Type", "video/mp2t")
	} else {
		response.Header().Add("Content-Type", "application/x-mpegURL")
	}

	response.Header().Add("Content-Length", fmt.Sprint(len(d)))
	response.Header().Add("Cache-Control", "max-age=31536000")

	response.WriteHeader(200)

	response.Write(d)
}

func api_handleAssetVideoPreviews(response http.ResponseWriter, request *http.Request) {
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

	asset_id, err := strconv.ParseUint(vars["asset"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
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
	asset_path := media.StartReadAsset(asset_id, ASSET_MUTI_FILE)

	s, err := CreateMultiFilePackReadStream(asset_path)

	if err != nil {
		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	if preview_img_index < 0 || preview_img_index >= s.file_count {
		s.Close()

		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(404)
		return
	}

	// Read raw data

	b, err := s.GetFile(preview_img_index)

	if err != nil {
		LogError(err)

		s.Close()

		media.EndReadAsset(asset_id)
		GetVault().media.ReleaseMediaResource(media_id)

		response.WriteHeader(500)
		return
	}

	s.Close()

	media.EndReadAsset(asset_id)
	GetVault().media.ReleaseMediaResource(media_id)

	// Decrypt the data

	d, err := decryptFileContents(b, session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	// Return the image

	response.Header().Add("Content-Type", "image/jpg")
	response.Header().Add("Content-Length", fmt.Sprint(len(d)))
	response.Header().Add("Cache-Control", "max-age=31536000")

	response.WriteHeader(200)

	response.Write(d)
}
