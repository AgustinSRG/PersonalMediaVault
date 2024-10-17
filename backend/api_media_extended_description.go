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

type ExtendedDescriptionSetBody struct {
	ExtendedDescription string `json:"ext_desc"`
}

type ExtendedDescriptionSetResponse struct {
	Url string `json:"url"`
}

func api_setExtendedDescription(response http.ResponseWriter, request *http.Request) {
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

	var p ExtendedDescriptionSetBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	assetData := []byte(p.ExtendedDescription)

	result := ThumbnailAPIResponse{
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
			os.Remove(desc_encrypted_file)
			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			os.Remove(desc_encrypted_file)

			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if meta == nil {
			media.CancelWrite()
			os.Remove(desc_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)
			ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
			return
		}

		desc_asset := meta.NextAssetID
		meta.NextAssetID++

		success, asset_path, asset_lock := media.AcquireAsset(desc_asset, ASSET_SINGLE_FILE)

		if !success {
			media.CancelWrite()
			os.Remove(desc_encrypted_file)
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
			os.Remove(desc_encrypted_file)
			GetVault().media.ReleaseMediaResource(media_id)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		// Change metadata

		if meta.HasExtendedDescription {
			// Remove old asset
			oldAsset := meta.ExtendedDescriptionAsset
			success, asset_path, asset_lock = media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

			if success {
				asset_lock.RequestWrite()
				asset_lock.StartWrite()

				os.Remove(asset_path)

				asset_lock.EndWrite()

				media.ReleaseAsset(oldAsset)
			}
		}

		meta.HasExtendedDescription = true
		meta.ExtendedDescriptionAsset = desc_asset
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(desc_asset) + "/ext_desc.txt" + "?fp=" + GetVault().credentials.GetFingerprint()

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

		if meta.HasExtendedDescription {
			// Remove old asset
			oldAsset := meta.ExtendedDescriptionAsset
			success, asset_path, asset_lock := media.AcquireAsset(oldAsset, ASSET_SINGLE_FILE)

			if success {
				asset_lock.RequestWrite()
				asset_lock.StartWrite()

				os.Remove(asset_path)

				asset_lock.EndWrite()

				media.ReleaseAsset(oldAsset)
			}
		}

		meta.HasExtendedDescription = false

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
