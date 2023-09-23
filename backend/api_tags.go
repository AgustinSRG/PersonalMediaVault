// Tags API

package main

import (
	"encoding/json"
	"net/http"
)

type TagListAPIItem struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func api_getTags(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	tag_list, err := GetVault().tags.ReadList(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	result := make([]TagListAPIItem, len(tag_list.Tags))

	i := 0

	for key, val := range tag_list.Tags {
		result[i] = TagListAPIItem{
			Id:   key,
			Name: val,
		}

		i++
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type TagAPISetBody struct {
	Media uint64 `json:"media_id"`
	Tag   string `json:"tag_name"`
}

func api_tagMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p TagAPISetBody

	err := json.NewDecoder(request.Body).Decode(&p)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Tag) == 0 || len(p.Tag) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid tag name provided")
		return
	}

	// Add to index

	tagName := ParseTagName(p.Tag)

	if len(tagName) == 0 || len(tagName) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid tag name provided")
		return
	}

	tag_id, err := GetVault().tags.TagMedia(p.Media, tagName, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Add to media metadata

	media := GetVault().media.AcquireMediaResource(p.Media)

	if media != nil {
		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(p.Media)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if meta != nil {
			meta.AddTag(tag_id)

			err = media.EndWrite(meta, session.key, false)

			GetVault().media.ReleaseMediaResource(p.Media)

			// Clear cache

			GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(p.Media)

			if err != nil {
				LogError(err)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		} else {
			media.CancelWrite()
			GetVault().media.ReleaseMediaResource(p.Media)
		}
	}

	var result TagListAPIItem

	result.Id = tag_id
	result.Name = tagName

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type UntagMediaBody struct {
	Media uint64 `json:"media_id"`
	Tag   uint64 `json:"tag_id"`
}

func api_untagMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.write {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p UntagMediaBody

	err := json.NewDecoder(request.Body).Decode(&p)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	err = GetVault().tags.UnTagMedia(p.Media, p.Tag, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Remove from media metadata

	media := GetVault().media.AcquireMediaResource(p.Media)

	if media != nil {
		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(p.Media)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if meta != nil {
			meta.RemoveTag(p.Tag)

			err = media.EndWrite(meta, session.key, false)

			GetVault().media.ReleaseMediaResource(p.Media)

			// Clear cache

			GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(p.Media)

			if err != nil {
				LogError(err)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		} else {
			media.CancelWrite()
			GetVault().media.ReleaseMediaResource(p.Media)
		}
	}

	response.WriteHeader(200)
}
