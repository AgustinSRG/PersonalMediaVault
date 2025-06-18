// Albums thumbnail assets API

// cSpell:ignore webp, nosniff

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
	"path/filepath"
	"strconv"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
	"github.com/gorilla/mux"
)

func api_handleAlbumThumbnailAssetGet(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)

	onlyHeader := request.Method == "HEAD"

	// Range header

	start, end := ParseRangeHeader(request)

	// Params

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
	ext := GetExtensionFromFileName(filename)
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
	}

	found, asset_path, asset_lock := GetVault().albums.AcquireThumbnailAsset(asset_id)

	if !found {
		response.WriteHeader(404)
		return
	}

	asset_lock.StartRead() // Start reading the asset

	s, err := encrypted_storage.CreateFileBlockEncryptReadStream(asset_path, session.key, FILE_PERMISSION)

	if err != nil {
		asset_lock.EndRead()
		GetVault().albums.ReleaseThumbnailAsset(asset_id)

		response.WriteHeader(404)
		return
	}

	fileSeek := int64(0)
	fileEnding := s.FileSize() - 1
	contentLength := s.FileSize()
	hasRange := false

	if start < 0 && end > 0 {
		// Only end
		fileSeek = s.FileSize() - end
		contentLength = end
		hasRange = true
	} else if start >= 0 && end < 0 {
		// Only start point
		fileSeek = start
		contentLength = s.FileSize() - start
		hasRange = true
	} else if start >= 0 && end > 0 {
		// Both start and end
		fileSeek = start
		fileEnding = end
		contentLength = end - start + 1
		hasRange = true
	}

	if fileSeek < 0 || fileSeek >= s.FileSize() || fileEnding >= s.FileSize() || fileEnding < fileSeek {
		// Invalid range
		s.Close()

		asset_lock.EndRead()
		GetVault().albums.ReleaseThumbnailAsset(asset_id)

		response.Header().Set("Content-Range", "bytes */"+fmt.Sprint(s.FileSize()))

		response.WriteHeader(416)
		return
	}

	if fileSeek > 0 {
		_, err := s.Seek(fileSeek, 0)
		if err != nil {
			// Seek error
			s.Close()

			asset_lock.EndRead()
			GetVault().albums.ReleaseThumbnailAsset(asset_id)

			LogError(err)
			response.WriteHeader(500)
			return
		}
	}

	// Send response

	response.Header().Set("Content-Type", mimeType)
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.Header().Set("Content-Length", fmt.Sprint(contentLength))
	response.Header().Set("Cache-Control", "max-age=31536000")

	if request.URL.Query().Get("download") == "force" {
		fileName := request.URL.Query().Get("filename")
		extPart := ext

		if len(extPart) > 0 {
			extPart = "." + extPart
		}

		if len(fileName) > 0 {
			response.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(fileName)+extPart+"\"")
		} else {
			response.Header().Set("Content-Disposition", "attachment")
		}
	}

	if hasRange {
		response.Header().Set("Content-Range", "bytes "+fmt.Sprint(fileSeek)+"-"+fmt.Sprint(fileEnding)+"/"+fmt.Sprint(s.FileSize()))
		response.WriteHeader(206)
	} else {
		response.Header().Set("Transfer-Encoding", "identity")
		response.WriteHeader(200)
	}

	if onlyHeader {
		s.Close()

		asset_lock.EndRead()
		GetVault().albums.ReleaseThumbnailAsset(asset_id)

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
			GetVault().albums.ReleaseThumbnailAsset(asset_id)

			return
		}

		if c <= 0 {
			s.Close()

			asset_lock.EndRead()
			GetVault().albums.ReleaseThumbnailAsset(asset_id)
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
			GetVault().albums.ReleaseThumbnailAsset(asset_id)

			return
		}
	}

	s.Close()

	asset_lock.EndRead()
	GetVault().albums.ReleaseThumbnailAsset(asset_id)
}

func api_editAlbumThumbnail(response http.ResponseWriter, request *http.Request) {
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

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

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

	// Probe uploaded file

	probe_data, err := ProbeMediaFileWithFFProbe(tempFile)

	if err != nil {
		LogError(err)

		DeleteTemporalFile(tempFile)

		ReturnAPIError(response, 400, "INVALID_THUMBNAIL", "Invalid thumbnail provided")
		return
	}

	// Generate a thumbnail

	thumbnail, err := GenerateThumbnailFromMedia(tempFile, probe_data)

	DeleteTemporalFile(tempFile)

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

	DeleteTemporalFile(thumbnail)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Get an ID for the asset

	thumb_asset, err := GetVault().albums.GetThumbnailAssetId(session.key)

	if err != nil {
		LogError(err)

		os.Remove(thumb_encrypted_file)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Save the asset

	success, asset_path, asset_lock := GetVault().albums.AcquireThumbnailAsset(thumb_asset)

	if !success {
		os.Remove(thumb_encrypted_file)

		ReturnAPIError(response, 404, "NOT_FOUND", "Album not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Ensure path exists
	parentPath := filepath.Dir(asset_path)
	_ = os.MkdirAll(parentPath, FOLDER_PERMISSION)

	// Move temp file
	err = RenameAndReplace(thumb_encrypted_file, asset_path)

	asset_lock.EndWrite()

	GetVault().albums.ReleaseThumbnailAsset(thumb_asset)

	if err != nil {
		LogError(err)

		os.Remove(asset_path)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change the thumbnail in the album metadata

	success, old_asset, err := GetVault().albums.SetAlbumThumbnail(album_id, thumb_asset, session.key)

	if err != nil {
		LogError(err)

		os.Remove(asset_path)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if !success {
		os.Remove(asset_path)

		ReturnAPIError(response, 404, "NOT_FOUND", "Album not found")
		return
	}

	// Remove old asset
	if old_asset != nil {
		success, asset_path, asset_lock := GetVault().albums.AcquireThumbnailAsset(*old_asset)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			GetVault().albums.ReleaseThumbnailAsset(*old_asset)
		}
	}

	// Response

	var result ThumbnailAPIResponse

	result.Url = "/assets/album_thumb/" + fmt.Sprint(thumb_asset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
