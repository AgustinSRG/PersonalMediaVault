// Albums API

package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type AlbumAPIItem struct {
	Id   uint64   `json:"id"`
	Name string   `json:"name"`
	List []uint64 `json:"list"`
}

func api_getAlbums(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
		return
	}

	albums, err := GetVault().albums.ReadAlbums(session.key)

	if err != nil {
		LogError(err)

		response.WriteHeader(500)
		return
	}

	result := make([]AlbumAPIItem, 0)

	for album_id, album := range albums.Albums {
		result = append(result, AlbumAPIItem{
			Id:   album_id,
			Name: album.Name,
			List: album.List,
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

type AlbumAPIDetail struct {
	Id   uint64              `json:"id"`
	Name string              `json:"name"`
	List []*MediaListAPIItem `json:"list"`
}

func api_getAlbum(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

		response.WriteHeader(500)
		return
	}

	album := albums.Albums[album_id]

	if album == nil {
		ReturnAPIError(response, 404, "NOT_FOUND", "Album not found")
		return
	}

	result := AlbumAPIDetail{
		Id:   album_id,
		Name: album.Name,
	}

	list := make([]*MediaListAPIItem, 0)

	for i := 0; i < len(album.List); i++ {
		list = append(list, GetMediaMinInfo(album.List[i], session))
	}

	result.List = list

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

type CreateAlbumAPIBody struct {
	Name string `json:"name"`
}

type CreateAlbumAPIResponse struct {
	Id uint64 `json:"album_id"`
}

func api_createAlbum(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

		response.WriteHeader(500)
		return
	}

	var result CreateAlbumAPIResponse

	result.Id = album_id

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

func api_deleteAlbum(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

		response.WriteHeader(500)
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
		response.WriteHeader(401)
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

		response.WriteHeader(500)
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
		response.WriteHeader(401)
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

		response.WriteHeader(500)
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
		response.WriteHeader(401)
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

		response.WriteHeader(500)
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
		response.WriteHeader(401)
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
	album_found, err := GetVault().albums.RemoveMediaFromAlbum(album_id, p.Id, session.key)

	if album_found {
		response.WriteHeader(200)
	} else {
		response.WriteHeader(404)
	}
}
