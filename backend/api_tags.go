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
		response.WriteHeader(401)
		return
	}

	tag_list, err := GetVault().tags.ReadList(session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	result := make([]TagListAPIItem, 0)

	for key, val := range tag_list.Tags {
		result = append(result, TagListAPIItem{
			Id:   key,
			Name: val,
		})
	}

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

type TagAPISetBody struct {
	Media uint64 `json:"media_id"`
	Tag   string `json:"tag_name"`
}

func api_tagMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

	tag_id, err := GetVault().tags.TagMedia(p.Media, tagName, session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	// Add to media metadata

	media := GetVault().media.AcquireMediaResource(p.Media)

	if media != nil {
		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(p.Media)

			response.WriteHeader(500)
			return
		}

		if meta != nil {
			meta.AddTag(tag_id)

			err = media.EndWrite(meta, session.key, false)

			GetVault().media.ReleaseMediaResource(p.Media)

			if err != nil {
				LogError(err)

				response.WriteHeader(500)
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

		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.Header().Add("Cache-Control", "no-cache")
	response.WriteHeader(200)

	response.Write(jsonResult)
}

type UntagMediaBody struct {
	Media uint64 `json:"media_id"`
	Tag   uint64 `json:"tag_id"`
}

func api_untagMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

		response.WriteHeader(500)
		return
	}

	// Remove from media metadata

	media := GetVault().media.AcquireMediaResource(p.Media)

	if media != nil {
		meta, err := media.StartWrite(session.key)

		if err != nil {
			LogError(err)

			GetVault().media.ReleaseMediaResource(p.Media)

			response.WriteHeader(500)
			return
		}

		if meta != nil {
			meta.RemoveTag(p.Tag)

			err = media.EndWrite(meta, session.key, false)

			GetVault().media.ReleaseMediaResource(p.Media)

			if err != nil {
				LogError(err)

				response.WriteHeader(500)
				return
			}
		} else {
			media.CancelWrite()
			GetVault().media.ReleaseMediaResource(p.Media)
		}
	}

	response.WriteHeader(200)
}
