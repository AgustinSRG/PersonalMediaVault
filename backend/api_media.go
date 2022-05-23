// Media API

package main

import (
	"fmt"
	"net/http"
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

func api_getMedia(response http.ResponseWriter, request *http.Request) {

}

func api_uploadMedia(response http.ResponseWriter, request *http.Request) {

}

func api_deleteMedia(response http.ResponseWriter, request *http.Request) {

}

func api_editMedia(response http.ResponseWriter, request *http.Request) {

}

func api_mediaRequestEncode(response http.ResponseWriter, request *http.Request) {

}

func api_mediaAddResolution(response http.ResponseWriter, request *http.Request) {

}

func api_mediaRemoveResolution(response http.ResponseWriter, request *http.Request) {

}
