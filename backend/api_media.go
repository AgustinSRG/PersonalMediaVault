// Media API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type MediaListAPIItem struct {
	Id          uint64    `json:"id"`
	Type        MediaType `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []uint64  `json:"tags"`
	Thumbnail   string    `json:"thumbnail"`
	Duration    float64   `json:"duration"`
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
			result.Thumbnail = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ThumbnailAsset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
		} else {
			result.Thumbnail = ""
		}
		result.Duration = meta.MediaDuration
		result.Description = meta.Description
		result.Tags = meta.Tags
	} else {
		result.Type = MediaTypeDeleted
		result.Title = ""
		result.Thumbnail = ""
		result.Duration = 0
		result.Description = ""
		result.Tags = make([]uint64, 0)
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

type MediaAPIMetaSubtitle struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type MediaAPIMetaAudio struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type MediaAPIMetaTimeSplit struct {
	Time float64 `json:"time"`
	Name string  `json:"name"`
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
	Fps      int32   `json:"fps"`

	Ready         bool   `json:"ready"`
	ReadyProgress int32  `json:"ready_p"`
	Encoded       bool   `json:"encoded"`
	Url           string `json:"url"`
	Task          uint64 `json:"task"`

	VideoPreviews         string  `json:"video_previews"`
	VideoPreviewsInterval float64 `json:"video_previews_interval"`

	Resolutions []MediaAPIMetaResolution `json:"resolutions"`

	Subtitles []MediaAPIMetaSubtitle `json:"subtitles"`

	Audios []MediaAPIMetaAudio `json:"audios"`

	ForceStartBeginning bool `json:"force_start_beginning"`

	TimeSlices []MediaAPIMetaTimeSplit `json:"time_slices"`

	HasImageNotes bool   `json:"img_notes"`
	ImageNotesURL string `json:"img_notes_url"`
}

func api_getMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
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
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.ReadMetadata(session.key)

	GetVault().media.ReleaseMediaResource(media_id)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
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

	result.Duration = meta.MediaDuration
	result.Width = meta.Width
	result.Height = meta.Height
	result.Fps = meta.Fps

	// Thumbnail

	if meta.ThumbnailReady {
		result.Thumbnail = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ThumbnailAsset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		result.Thumbnail = ""
	}

	// Original

	if meta.OriginalReady && meta.OriginalEncoded {
		result.Ready = true
		result.Encoded = true
		result.Task = meta.OriginalTask
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.OriginalAsset) + "/original_" + fmt.Sprint(media_id) + "." + meta.OriginalExtension + "?fp=" + GetVault().credentials.GetFingerprint()
	} else if meta.OriginalReady {
		result.Ready = true
		result.Encoded = false
		result.Task = meta.OriginalTask
		result.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.OriginalAsset) + "/original_" + fmt.Sprint(media_id) + "." + meta.OriginalExtension + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		result.Ready = false
		result.ReadyProgress = GetVault().media.GetProgress(media_id)
		result.Encoded = false
		result.Task = meta.OriginalTask
		result.Url = ""
	}

	// Video previews

	if meta.PreviewsReady {
		result.VideoPreviewsInterval = meta.PreviewsInterval
		result.VideoPreviews = "/assets/p/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.PreviewsAsset) + "/preview_{INDEX}.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		result.VideoPreviewsInterval = 0
		result.VideoPreviews = ""
	}

	// Image notes

	if meta.HasImageNotes {
		result.HasImageNotes = true
		result.ImageNotesURL = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.ImageNotesAsset) + "/notes.json" + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		result.HasImageNotes = false
		result.ImageNotesURL = ""
	}

	// Resolutions

	var resolutions []MediaAPIMetaResolution

	if meta.Resolutions != nil {
		resolutions = make([]MediaAPIMetaResolution, len(meta.Resolutions))
		for i := 0; i < len(meta.Resolutions); i++ {
			var r MediaAPIMetaResolution

			r.Width = meta.Resolutions[i].Width
			r.Height = meta.Resolutions[i].Height
			r.Fps = meta.Resolutions[i].Fps

			if meta.Resolutions[i].Ready {
				r.Ready = true
				r.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.Resolutions[i].Asset) + "/video." + meta.Resolutions[i].Extension + "?fp=" + GetVault().credentials.GetFingerprint()
			} else {
				r.Ready = false
				r.Url = ""
				r.Task = meta.Resolutions[i].TaskId
			}

			resolutions[i] = r
		}
	} else {
		resolutions = make([]MediaAPIMetaResolution, 0)
	}

	sort.Slice(resolutions, func(i, j int) bool {
		areaI := resolutions[i].Width * resolutions[i].Height
		areaJ := resolutions[j].Width * resolutions[j].Height
		if areaI > areaJ {
			return true
		} else if areaI < areaJ {
			return false
		} else {
			return resolutions[i].Fps > resolutions[j].Fps
		}
	})

	result.Resolutions = resolutions

	// Subtitles

	var subtitles []MediaAPIMetaSubtitle

	if meta.Subtitles != nil {
		subtitles = make([]MediaAPIMetaSubtitle, len(meta.Subtitles))

		for i := 0; i < len(meta.Subtitles); i++ {
			var s MediaAPIMetaSubtitle

			s.Id = meta.Subtitles[i].Id
			s.Name = meta.Subtitles[i].Name

			s.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.Subtitles[i].Asset) + "/subrip.srt" + "?fp=" + GetVault().credentials.GetFingerprint()

			subtitles[i] = s
		}
	} else {
		subtitles = make([]MediaAPIMetaSubtitle, 0)
	}

	result.Subtitles = subtitles

	// Audios

	var audios []MediaAPIMetaAudio

	if meta.Subtitles != nil {
		audios = make([]MediaAPIMetaAudio, len(meta.AudioTracks))

		for i := 0; i < len(meta.AudioTracks); i++ {
			var s MediaAPIMetaAudio

			s.Id = meta.AudioTracks[i].Id
			s.Name = meta.AudioTracks[i].Name

			s.Url = "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(meta.AudioTracks[i].Asset) + "/audio.mp3" + "?fp=" + GetVault().credentials.GetFingerprint()

			audios[i] = s
		}
	} else {
		audios = make([]MediaAPIMetaAudio, 0)
	}

	result.Audios = audios

	// Extra

	result.ForceStartBeginning = meta.ForceStartBeginning

	// Time slices

	var timeSlices []MediaAPIMetaTimeSplit

	if meta.Splits != nil {
		timeSlices = make([]MediaAPIMetaTimeSplit, len(meta.Splits))

		for i := 0; i < len(meta.Splits); i++ {
			var s MediaAPIMetaTimeSplit

			s.Time = meta.Splits[i].Time
			s.Name = meta.Splits[i].Name

			timeSlices[i] = s
		}
	} else {
		timeSlices = make([]MediaAPIMetaTimeSplit, 0)
	}

	result.TimeSlices = timeSlices

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type AssetSizeAPIResponse struct {
	Id   uint64 `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type MediaSizeAPIResponse struct {
	MetaSize  int64                  `json:"meta_size"`
	AssetSize []AssetSizeAPIResponse `json:"assets"`
}

func api_getMediaSizeStats(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
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
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.ReadMetadata(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	metaSize, err := media.GetMetadataSize()

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	var result MediaSizeAPIResponse

	result.MetaSize = metaSize
	result.AssetSize = make([]AssetSizeAPIResponse, 0)

	if meta.OriginalReady {
		result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
			Id:   meta.OriginalAsset,
			Type: ASSET_SINGLE_FILE,
			Name: "ORIGINAL",
			Size: 0,
		})
	}

	if meta.ThumbnailReady {
		result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
			Id:   meta.ThumbnailAsset,
			Type: ASSET_SINGLE_FILE,
			Name: "THUMBNAIL",
			Size: 0,
		})
	}

	if meta.PreviewsReady {
		result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
			Id:   meta.PreviewsAsset,
			Type: ASSET_MULTI_FILE,
			Name: "VIDEO_PREVIEWS",
			Size: 0,
		})
	}

	if meta.Resolutions != nil {
		for i := 0; i < len(meta.Resolutions); i++ {
			result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
				Id:   meta.Resolutions[i].Asset,
				Type: ASSET_SINGLE_FILE,
				Name: fmt.Sprintf("RESIZED_%vx%v:%v", meta.Resolutions[i].Width, meta.Resolutions[i].Height, meta.Resolutions[i].Fps),
				Size: 0,
			})
		}
	}

	if meta.Subtitles != nil {
		for i := 0; i < len(meta.Subtitles); i++ {
			result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
				Id:   meta.Subtitles[i].Asset,
				Type: ASSET_SINGLE_FILE,
				Name: fmt.Sprintf("SUBTITLES_%v", strings.ToUpper(meta.Subtitles[i].Id)),
				Size: 0,
			})
		}
	}

	if meta.HasImageNotes {
		result.AssetSize = append(result.AssetSize, AssetSizeAPIResponse{
			Id:   meta.ImageNotesAsset,
			Type: ASSET_SINGLE_FILE,
			Name: "IMG_NOTES",
			Size: 0,
		})
	}

	for i := 0; i < len(result.AssetSize); i++ {
		found, asset_path, asset_lock := media.AcquireAsset(result.AssetSize[i].Id, result.AssetSize[i].Type)

		if !found {
			continue
		}

		asset_lock.StartRead() // Start reading the asset

		stats, err := os.Stat(asset_path)

		if err != nil {
			asset_lock.EndRead()
			continue
		}

		asset_lock.EndRead()

		result.AssetSize[i].Size = stats.Size()
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

func api_getMediaAlbums(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	albumsData, err := GetVault().albums.readData(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	albums := make([]uint64, 0)

	if albumsData.Albums != nil {
		for id, album := range albumsData.Albums {
			if album.HasMedia(media_id) {
				albums = append(albums, id)
			}
		}
	}

	// Response

	jsonResult, err := json.Marshal(albums)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type MediaAPIEditTitleBody struct {
	Title string `json:"title"`
}

func api_editMediaTitle(response http.ResponseWriter, request *http.Request) {
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

	var p MediaAPIEditTitleBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Title) == 0 || len(p.Title) > 255 {
		ReturnAPIError(response, 400, "INVALID_TITLE", "Invalid title provided")
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

	meta.Title = p.Title

	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

type MediaAPIEditDescriptionBody struct {
	Description string `json:"description"`
}

func api_editMediaDescription(response http.ResponseWriter, request *http.Request) {
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

	var p MediaAPIEditDescriptionBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Description) > (5 * 1024 * 1024) {
		ReturnAPIError(response, 400, "INVALID_DESCRIPTION", "Invalid description provided")
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

	meta.Description = p.Description

	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

type MediaAPIEditExtraParams struct {
	ForceStartBeginning bool `json:"force_start_beginning"`
}

func api_editMediaExtraParams(response http.ResponseWriter, request *http.Request) {
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

	var p MediaAPIEditExtraParams

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
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

	meta.ForceStartBeginning = p.ForceStartBeginning

	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

func api_editMediaTimelineSplices(response http.ResponseWriter, request *http.Request) {
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

	var p []MediaAPIMetaTimeSplit

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p == nil {
		response.WriteHeader(400)
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

	timeSplits := make([]MediaSplit, 0)

	for i := 0; i < len(p) && i < 1024; i++ {
		timeSplits = append(timeSplits, MediaSplit{
			Time: p[i].Time,
			Name: LimitStringSize(p[i].Name, 80),
		})
	}

	meta.Splits = timeSplits

	err = media.EndWrite(meta, session.key, false)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

func api_mediaRequestEncode(response http.ResponseWriter, request *http.Request) {
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

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

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

	// Check for unattended encoded assets

	// Check original

	if meta.OriginalReady && !meta.OriginalEncoded {
		// Check task

		task_info := GetVault().tasks.GetTaskInfo(meta.OriginalTask)

		if task_info == nil {
			// Task crashed or was never spawned, make a new one

			meta.OriginalTask = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_ORIGINAL, nil, false)
		}
	} else if meta.OriginalReady && meta.OriginalEncoded {
		meta.OriginalEncoded = false
		meta.OriginalTask = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_ORIGINAL, nil, false)
	}

	// Check previews

	if meta.Type == MediaTypeVideo && !meta.PreviewsReady {
		// Check task

		task_info := GetVault().tasks.GetTaskInfo(meta.PreviewsTask)

		if task_info == nil {
			// Task crashed or was never spawned, make a new one

			meta.PreviewsTask = GetVault().tasks.AddTask(session, media_id, TASK_IMAGE_PREVIEWS, nil, false)
		}
	}

	// Check resolutions

	if meta.Resolutions != nil {
		for i := 0; i < len(meta.Resolutions); i++ {
			if !meta.Resolutions[i].Ready {
				task_info := GetVault().tasks.GetTaskInfo(meta.Resolutions[i].TaskId)

				if task_info == nil {
					// Task crashed or was never spawned, make a new one

					meta.Resolutions[i].TaskId = GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
						Width:  meta.Resolutions[i].Width,
						Height: meta.Resolutions[i].Height,
						Fps:    meta.Resolutions[i].Fps,
					}, false)
				}
			}
		}
	}

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

type ApiMediaResolutionBody struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	Fps    int32 `json:"fps"`
}

func api_mediaAddResolution(response http.ResponseWriter, request *http.Request) {
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

	var p ApiMediaResolutionBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Width <= 0 || p.Height <= 0 {
		response.WriteHeader(400)
		return
	}

	if p.Fps <= 0 {
		p.Fps = 1
	}

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["mid"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

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

	if meta.Type != MediaTypeVideo && meta.Type != MediaTypeImage {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "NOT_SUPPORTED", "This feature is not supported for the media type. Only for videos and images.")
		return
	}

	if meta.Type == MediaTypeImage {
		p.Fps = 1
	}

	if meta.Resolutions == nil {
		meta.Resolutions = make([]MediaResolution, 0)
	}

	if meta.FindResolution(p.Width, p.Height, p.Fps) != -1 {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)
		ReturnAPIError(response, 400, "DUPLICATED_RESOLUTION", "Duplicated resolution.")
		return
	}

	// Spawn task

	task_id := GetVault().tasks.AddTask(session, media_id, TASK_ENCODE_RESOLUTION, &UserConfigResolution{
		Width:  p.Width,
		Height: p.Height,
		Fps:    p.Fps,
	}, false)

	// Save resolution

	resolution := MediaResolution{
		Width:     p.Width,
		Height:    p.Height,
		Fps:       p.Fps,
		Ready:     false,
		Asset:     0,
		Extension: "",
		TaskId:    task_id,
	}

	meta.Resolutions = append(meta.Resolutions, resolution)

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	var result MediaAPIMetaResolution

	result.Width = p.Width
	result.Height = p.Height
	result.Fps = p.Fps
	result.Ready = false
	result.Task = task_id
	result.Url = ""

	// Response

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_mediaRemoveResolution(response http.ResponseWriter, request *http.Request) {
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

	var p ApiMediaResolutionBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
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
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.StartWriteWithFullLock(session.key)

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

	if meta.Resolutions == nil {
		meta.Resolutions = make([]MediaResolution, 0)
	}

	res_index := meta.FindResolution(p.Width, p.Height, p.Fps)

	if res_index >= 0 {
		r := meta.Resolutions[res_index]

		if r.Ready {
			// Remove asset
			success, asset_path, asset_lock := media.AcquireAsset(r.Asset, ASSET_SINGLE_FILE)

			if success {
				asset_lock.RequestWrite()
				asset_lock.StartWrite()

				os.Remove(asset_path)

				asset_lock.EndWrite()

				media.ReleaseAsset(r.Asset)
			}
		} else {
			// Kill Task
			GetVault().tasks.KillTask(r.TaskId)
		}

		// Remove resolution from metadata
		meta.RemoveResolution(res_index)
	}

	err = media.EndWrite(meta, session.key, true)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().media.ReleaseMediaResource(media_id)

	response.WriteHeader(200)
}

func api_deleteMedia(response http.ResponseWriter, request *http.Request) {
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

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	meta, err := media.ReadMetadata(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if meta == nil {
		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 404, "NOT_FOUND", "Media not found")
		return
	}

	// Untag

	if meta.Tags != nil {
		for i := 0; i < len(meta.Tags); i++ {
			err = GetVault().tags.UnTagMedia(media_id, meta.Tags[i], session.key)

			if err != nil {
				LogError(err)

				GetVault().media.ReleaseMediaResource(media_id)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		}
	}

	// Remove from main index
	main_index, err := GetVault().index.StartWrite()

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	_, _, err = main_index.file.RemoveValue(media_id)

	if err != nil {
		LogError(err)

		GetVault().index.CancelWrite(main_index)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	err = GetVault().index.EndWrite(main_index)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Remove from albums

	err = GetVault().albums.OnMediaDelete(media_id, session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Delete
	media.Delete()

	GetVault().media.ReleaseMediaResource(media_id)

	// Kill any task pending for that media
	GetVault().tasks.KillTaskByMedia(media_id)

	response.WriteHeader(200)
}
