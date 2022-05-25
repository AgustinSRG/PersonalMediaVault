// Media API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type MediaListAPIItem struct {
	Id        uint64    `json:"id"`
	Type      MediaType `json:"type"`
	Title     string    `json:"title"`
	Thumbnail string    `json:"thumbnail"`
	Duration  float64   `json:"duration"`
	Width     int32     `json:"width"`
	Height    int32     `json:"height"`
}

func GetMediaMinInfo(media_id uint64, session *ActiveSession) *MediaListAPIItem {
	var result MediaListAPIItem

	result.Id = media_id

	media := GetVault().media.AcquireMediaResource(media_id)

	meta, err := media.ReadMetadata(session.key)

	if err != nil {
		LogError(err)
	}

	if meta != nil {
		result.Type = meta.Type
		result.Title = meta.Title
		if meta.ThumbnailReady {
			result.Thumbnail = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ThumbnailAsset) + "/thumbnail.jpg"
		} else {
			result.Thumbnail = ""
		}
		result.Duration = meta.MediaDuration
		result.Width = meta.Width
		result.Height = meta.Height
	} else {
		result.Type = MediaTypeDeleted
		result.Title = ""
		result.Thumbnail = ""
		result.Duration = 0
		result.Width = 0
		result.Height = 0
	}

	GetVault().media.ReleaseMediaResource(media_id)

	return &result
}

type MediaAPIMetaResolution struct {
	Width  int32  `json:"width"`
	Height int32  `json:"height"`
	Fps    int32  `json:"fps"`
	Ready  bool   `json:"ready"`
	Url    string `json:"url"`
	Task   uint64 `json:"task"`
}

type MediaAPIMetaResponse struct {
	Id   uint64    `json:"id"`
	Type MediaType `json:"type"`

	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Tags            []uint64 `json:"tags"`
	UploadTimestamp int64    `json:"upload_time"`

	Thumbnail string `json:"thumbnail"`

	Duration float64 `json:"duration"`
	Width    int32   `json:"width"`
	Height   int32   `json:"height"`

	Ready   bool   `json:"ready"`
	Encoded bool   `json:"encoded"`
	Url     string `json:"url"`
	Task    uint64 `json:"task"`

	VideoPreviews         string  `json:"video_previews"`
	VideoPreviewsInterval float64 `json:"video_previews_interval"`

	Resolutions []MediaAPIMetaResolution `json:"resolutions"`
}

func api_getMedia(response http.ResponseWriter, request *http.Request) {
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

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 400, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.ReadMetadata(session.key)

	GetVault().media.ReleaseMediaResource(media_id)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	if meta == nil {
		ReturnAPIError(response, 400, "NOT_FOUND", "Media not found")
		return
	}

	var result MediaAPIMetaResponse

	// Set result

	result.Id = media_id
	result.Type = meta.Type

	result.Title = meta.Title
	result.Description = meta.Description
	result.Tags = meta.Tags
	result.UploadTimestamp = meta.UploadTimestamp

	if meta.ThumbnailReady {
		result.Thumbnail = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ThumbnailAsset) + "/thumbnail.jpg"
	} else {
		result.Thumbnail = ""
	}

	result.Duration = meta.MediaDuration
	result.Width = meta.Width
	result.Height = meta.Height

	if meta.OriginalReady && meta.OriginalEncoded {
		result.Ready = true
		result.Encoded = true
		result.Task = meta.OriginalTask
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.OriginalAsset) + "/video." + meta.OriginalExtension
	} else if meta.OriginalReady {
		result.Ready = true
		result.Encoded = false
		result.Task = meta.OriginalTask
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.OriginalAsset) + "/video." + meta.OriginalExtension
	} else {
		result.Ready = false
		result.Encoded = false
		result.Task = meta.OriginalTask
		result.Url = ""
	}

	if meta.PreviewsReady {
		result.VideoPreviewsInterval = meta.PreviewsInterval
		result.VideoPreviews = "/assets/p/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.PreviewsAsset) + "/preview_{INDEX}.jpg"
	} else {
		result.VideoPreviewsInterval = 0
		result.VideoPreviews = ""
	}

	resolutions := make([]MediaAPIMetaResolution, 0)

	if meta.Resolutions != nil {
		for i := 0; i < len(meta.Resolutions); i++ {
			var r MediaAPIMetaResolution

			r.Width = meta.Resolutions[i].Width
			r.Height = meta.Resolutions[i].Height
			r.Fps = meta.Resolutions[i].Fps

			if meta.Resolutions[i].Ready {
				r.Ready = true
				r.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.Resolutions[i].Asset) + "/video." + meta.Resolutions[i].Extension
			} else {
				r.Ready = false
				r.Url = ""
				r.Task = meta.Resolutions[i].TaskId
			}

			resolutions = append(resolutions, r)
		}
	}

	result.Resolutions = resolutions

	// Response

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

func api_editMediaTitle(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_editMediaDescription(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_editMediaThumbnail(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_mediaRequestEncode(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_mediaAddResolution(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_mediaRemoveResolution(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}

func api_deleteMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}
}
