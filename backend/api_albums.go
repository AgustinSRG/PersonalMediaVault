// Albums API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AlbumAPIItem struct {
	Id           uint64 `json:"id"`
	Name         string `json:"name"`
	Size         int    `json:"size"`
	Thumbnail    string `json:"thumbnail"` // This thumbnail is from the first media asset in the album
	LastModified int64  `json:"lm"`        // Last modified timestamp
}

type AlbumAPIItemMinified struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func getAlbumThumbnail(album_id uint64, album *VaultAlbumData, session *ActiveSession) string {
	if album.List == nil || len(album.List) == 0 {
		return ""
	}

	media_id := album.List[0]

	ok, has_thumbnail, thumbnail_asset := GetVault().albums.FindCachedThumbnailAsset(album_id)

	if ok {
		if has_thumbnail {
			return "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(thumbnail_asset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
		} else {
			return ""
		}
	}

	media := GetVault().media.AcquireMediaResource(media_id)

	meta, err := media.ReadMetadata(session.key)

	GetVault().media.ReleaseMediaResource(media_id)

	if err != nil {
		LogError(err)
	}

	if meta == nil {
		return ""
	}

	has_thumbnail = meta.ThumbnailReady
	thumbnail_asset = meta.ThumbnailAsset

	GetVault().albums.thumbnail_cache[album_id] = ThumbnailCacheEntry{
		has_thumbnail: has_thumbnail,
		asset:         thumbnail_asset,
	}

	if has_thumbnail {
		return "/assets/b/" + fmt.Sprint(media_id) + "/" + fmt.Sprint(thumbnail_asset) + "/thumbnail.jpg" + "?fp=" + GetVault().credentials.GetFingerprint()
	} else {
		return ""
	}
}

func api_getAlbums(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	minMode := request.URL.Query().Get("mode") == "min"

	if !minMode {
		GetVault().albums.thumbnail_cache_lock.Lock()
	}

	albums, err := GetVault().albums.ReadAlbums(session.key)

	if err != nil {
		if !minMode {
			GetVault().albums.thumbnail_cache_lock.Unlock()
		}

		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if !minMode {
		result := make([]AlbumAPIItem, len(albums.Albums))

		i := 0

		for album_id, album := range albums.Albums {

			result[i] = AlbumAPIItem{
				Id:           album_id,
				Name:         album.Name,
				Size:         len(album.List),
				Thumbnail:    getAlbumThumbnail(album_id, album, session),
				LastModified: album.LastModified,
			}

			i++
		}

		GetVault().albums.thumbnail_cache_lock.Unlock()

		jsonResult, err := json.Marshal(result)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		ReturnAPI_JSON(response, request, jsonResult)
	} else {
		result := make([]AlbumAPIItemMinified, len(albums.Albums))

		i := 0

		for album_id, album := range albums.Albums {
			result[i] = AlbumAPIItemMinified{
				Id:   album_id,
				Name: album.Name,
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
}

type AlbumAPIDetail struct {
	Id           uint64              `json:"id"`
	Name         string              `json:"name"`
	List         []*MediaListAPIItem `json:"list"`
	LastModified int64               `json:"lm"` // Last modified timestamp
}

func api_getAlbum(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	vars := mux.Vars(request)

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	albums, err := GetVault().albums.ReadAlbums(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	album := albums.Albums[album_id]

	if album == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Album not found")
		return
	}

	result := AlbumAPIDetail{
		Id:           album_id,
		Name:         album.Name,
		LastModified: album.LastModified,
	}

	list := make([]*MediaListAPIItem, len(album.List))

	for i := 0; i < len(album.List); i++ {
		list[i] = GetMediaMinInfo(album.List[i], session)
	}

	result.List = list

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type CreateAlbumAPIBody struct {
	Name string `json:"name"`
}

type CreateAlbumAPIResponse struct {
	Id uint64 `json:"album_id"`
}

func api_createAlbum(response http.ResponseWriter, request *http.Request) {
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

	var p RenameAlbumAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Name) == 0 || len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid album name provided")
		return
	}

	album_id, err := GetVault().albums.CreateAlbum(p.Name, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	var result CreateAlbumAPIResponse

	result.Id = album_id

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_deleteAlbum(response http.ResponseWriter, request *http.Request) {
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

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	err = GetVault().albums.DeleteAlbum(album_id, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	response.WriteHeader(200)
}

type RenameAlbumAPIBody struct {
	Name string `json:"name"`
}

func api_renameAlbum(response http.ResponseWriter, request *http.Request) {
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

	var p RenameAlbumAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Name) == 0 || len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "Invalid album name provided")
		return
	}

	vars := mux.Vars(request)

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Start Write
	album_found, err := GetVault().albums.RenameAlbum(album_id, p.Name, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if album_found {
		response.WriteHeader(200)
	} else {
		response.WriteHeader(404)
	}
}

type AlbumSetListAPIBody struct {
	List []uint64 `json:"list"`
}

func api_setAlbumList(response http.ResponseWriter, request *http.Request) {
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

	var p AlbumSetListAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.List == nil {
		ReturnAPIError(response, 400, "INVALID_LIST", "Invalid list provided")
		return
	}

	vars := mux.Vars(request)

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Start Write
	album_found, err := GetVault().albums.SetAlbumList(album_id, p.List, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if album_found {
		response.WriteHeader(200)
	} else {
		response.WriteHeader(404)
	}
}

type AlbumMediaAPIBody struct {
	Id uint64 `json:"media_id"`
}

func api_albumAddMedia(response http.ResponseWriter, request *http.Request) {
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

	var p AlbumMediaAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	vars := mux.Vars(request)

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Start Write
	album_found, err := GetVault().albums.AddMediaToAlbum(album_id, p.Id, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if album_found {
		response.WriteHeader(200)
	} else {
		response.WriteHeader(404)
	}
}

func api_albumRemoveMedia(response http.ResponseWriter, request *http.Request) {
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

	var p AlbumMediaAPIBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	vars := mux.Vars(request)

	album_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	album_found, err := GetVault().albums.RemoveMediaFromAlbum(album_id, p.Id, session.key)

	if err != nil {
		LogError(err)
		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if album_found {
		response.WriteHeader(200)
	} else {
		response.WriteHeader(404)
	}
}
