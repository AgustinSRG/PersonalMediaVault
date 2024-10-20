// Image notes API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

type ImageNote struct {
	XPosition int    `json:"x"`
	YPosition int    `json:"y"`
	Width     int    `json:"w"`
	Height    int    `json:"h"`
	Text      string `json:"text"`
}

type ImageNotesSetResponse struct {
	Url string `json:"url"`
}

func api_setImageNotes(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	var p []ImageNote = make([]ImageNote, 0)

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	assetData, err := json.Marshal(p)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	result := ImageNotesSetResponse{
		Url: "",
	}

	// Encrypt the notes file

	notes_encrypted_file, err := EncryptAssetData(assetData, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Put the notes into the media assets

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		os.Remove(notes_encrypted_file)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		os.Remove(notes_encrypted_file)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		media.CancelWrite()
		os.Remove(notes_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	if meta.Type != MediaTypeImage {
		media.CancelWrite()
		os.Remove(notes_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "NOT_SUPPORTED", "This feature is not supported for the media type. Only for images.")
		return
	}

	notes_asset := meta.NextAssetID
	meta.NextAssetID++

	success, asset_path, asset_lock := media.AcquireAsset(notes_asset, ASSET_SINGLE_FILE)

	if !success {
		media.CancelWrite()
		os.Remove(notes_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(notes_encrypted_file, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(notes_asset)

	if err != nil {
		LogError(err)

		media.CancelWrite()
		os.Remove(notes_encrypted_file)
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Change metadata

	if meta.HasImageNotes {
		// Remove old asset
		oldAsset := meta.ImageNotesAsset
		success, asset_path, asset_lock = media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

		if success {
			asset_lock.RequestWrite()
			asset_lock.StartWrite()

			os.Remove(asset_path)

			asset_lock.EndWrite()

			media.ReleaseAsset(oldAsset)
		}
	}

	meta.HasImageNotes = true
	meta.ImageNotesAsset = notes_asset
	result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(notes_asset) + "/notes.json" + "?fp=" + GetVault().credentials.GetFingerprint()

	// Save
	err = media.EndWrite(meta, session.key, false)

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
