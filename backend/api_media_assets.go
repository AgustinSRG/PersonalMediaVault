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

}

func api_handleAssetVideoHLS(response http.ResponseWriter, request *http.Request) {

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
