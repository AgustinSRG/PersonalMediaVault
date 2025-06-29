// Home page API

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Home page group data (for the API)
type HomePageGroupApi struct {
	// Group unique ID
	Id uint64 `json:"id"`

	// Type of group
	Type uint8 `json:"type"`

	// Name of the group, in order to display it
	Name string `json:"name"`

	// List of elements of the group
	ElementsCount int `json:"elementsCount"`
}

func minifyHomePageGroup(g *HomePageGroup) HomePageGroupApi {
	return HomePageGroupApi{
		Id:            g.Id,
		Type:          g.Type,
		Name:          g.Name,
		ElementsCount: len(g.Elements),
	}
}

func api_listHomePageGroups(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	res := make([]HomePageGroupApi, len(config.Groups))

	for i, g := range config.Groups {
		res[i] = minifyHomePageGroup(&g)
	}

	jsonResult, err := json.Marshal(res)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type HomeElementApiResult struct {
	media *MediaListAPIItem
	album *AlbumAPIItem
}

// Gets media minified info (preview) for a single item in a list
// Runs in a co-routine
func GetMediaMinInfoListTaskHome(mediaId uint64, session *ActiveSession, result []*HomeElementApiResult, index int, wg *sync.WaitGroup) {
	defer wg.Done()
	result[index] = &HomeElementApiResult{
		media: GetMediaMinInfo(mediaId, session),
	}
}

// Gets media minified info (preview) for a list
// Uses co-routines
func GetHomeElementsMinInfoList(list []HomePageElement, albums *VaultAlbumsData, session *ActiveSession) []*HomeElementApiResult {
	result := make([]*HomeElementApiResult, len(list))

	if len(list) == 0 {
		return result
	}

	wg := &sync.WaitGroup{}

	wg.Add(len(list))

	for i, e := range list {
		if e.ElementType == HOME_PAGE_ELEMENT_TYPE_ALBUM {
			album := albums.Albums[e.Id]

			if album != nil {
				result[i] = &HomeElementApiResult{
					album: &AlbumAPIItem{
						Id:           e.Id,
						Name:         album.Name,
						Size:         len(album.List),
						Thumbnail:    getAlbumThumbnail(e.Id, album, session),
						LastModified: album.LastModified,
					},
				}
			} else {
				result[i] = &HomeElementApiResult{}
			}
		} else {
			go GetMediaMinInfoListTaskHome(e.Id, session, result, i, wg)
		}
	}

	wg.Wait()

	return result
}

func api_getHomePageGroupElements(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	vars := mux.Vars(request)

	group_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	groupPos := config.FindGroup(group_id)

	if groupPos == -1 {
		ReturnAPIError(response, 404, "GROUP_NOT_FOUND", "No group was found with the specified ID.")
		return
	}

	elements := config.Groups[groupPos].Elements

	albums, err := GetVault().albums.ReadAlbums(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	result := GetHomeElementsMinInfoList(elements, albums, session)

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func api_createHomePageGroup(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}
}

func api_renameHomePageGroup(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}
}

func api_moveHomePageGroup(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}
}

func api_setElementsHomePageGroup(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}
}

func api_deleteHomePageGroup(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}
}
