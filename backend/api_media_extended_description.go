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

type MediaAPIEditDescriptionBody struct {
	Description string `json:"description"`
}

type DescriptionSetResponse struct {
	Url string `json:"url"`
}

func api_setDescription(response http.ResponseWriter, request *http.Request) {
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

	var p MediaAPIEditDescriptionBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	assetData := []byte(p.Description)

	result := DescriptionSetResponse{
		Url: "",
	}

	if len(assetData) > 0 {
		// Encrypt the description file

		desc_encrypted_file, err := EncryptAssetData(assetData, session.key)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		// Put the description into the media assets

		media := GetVault().media.AcquireMediaResource(media_id)

		if media == nil {
			_ = os.Remove(desc_encrypted_file)

			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			_ = os.Remove(desc_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if meta == nil {
			media.CancelWrite()

			_ = os.Remove(desc_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		desc_asset := meta.NextAssetID
		meta.NextAssetID++

		success, asset_path, asset_lock := media.AcquireAsset(desc_asset, ASSET_SINGLE_FILE)

		if !success {
			media.CancelWrite()

			_ = os.Remove(desc_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		asset_lock.RequestWrite()
		asset_lock.StartWrite()

		// Move temp file
		err = RenameAndReplace(desc_encrypted_file, asset_path)

		asset_lock.EndWrite()

		media.ReleaseAsset(desc_asset)

		if err != nil {
			LogError(err)

			media.CancelWrite()

			_ = os.Remove(desc_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		// Change metadata

		if meta.HasDescription {
			// Remove old asset
			oldAsset := meta.DescriptionAsset
			success, asset_path, asset_lock = media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

			if success {
				asset_lock.RequestWrite()
				asset_lock.StartWrite()

				_ = os.Remove(asset_path)

				asset_lock.EndWrite()

				media.ReleaseAsset(oldAsset)
			}
		}

		meta.HasDescription = true
		meta.DescriptionAsset = desc_asset
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(desc_asset) + "/description.txt" + "?fp=" + GetVault().credentials.GetFingerprint()

		// Save
		err = media.EndWrite(meta, session.key, false)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		GetVault().media.ReleaseMediaResource(media_id)
	} else {
		// Put the description into the media assets

		media := GetVault().media.AcquireMediaResource(media_id)

		if media == nil {
			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if meta == nil {
			media.CancelWrite()
			GetVault().media.ReleaseMediaResource(media_id)
			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		// Change metadata

		if meta.HasDescription {
			// Remove old asset
			oldAsset := meta.DescriptionAsset
			success, asset_path, asset_lock := media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

			if success {
				asset_lock.RequestWrite()
				asset_lock.StartWrite()

				_ = os.Remove(asset_path)

				asset_lock.EndWrite()

				media.ReleaseAsset(oldAsset)
			}
		}

		meta.HasDescription = false

		// Save
		err = media.EndWrite(meta, session.key, false)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		GetVault().media.ReleaseMediaResource(media_id)
	}

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
