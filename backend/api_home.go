// Home page API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// Max number of elements in a group
const GROUP_ELEMENTS_LIMIT = 256

// Max number of groups
const GROUPS_COUNT_LIMIT = 1024

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
	Media *MediaListAPIItem `json:"media,omitempty"`
	Album *AlbumAPIItem     `json:"album,omitempty"`
}

// Gets media minified info (preview) for a single item in a list
// Runs in a co-routine
func GetMediaMinInfoListTaskHome(mediaId uint64, session *ActiveSession, result []*HomeElementApiResult, index int, wg *sync.WaitGroup) {
	defer wg.Done()
	result[index] = &HomeElementApiResult{
		Media: GetMediaMinInfo(mediaId, session, nil),
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
					Album: &AlbumAPIItem{
						Id:           e.Id,
						Name:         album.Name,
						Size:         len(album.List),
						Thumbnail:    getAlbumThumbnail(e.Id, album, session),
						LastModified: album.LastModified,
					},
				}
			} else {
				result[i] = &HomeElementApiResult{
					Album: &AlbumAPIItem{
						Id:           e.Id,
						Name:         "",
						Size:         0,
						Thumbnail:    "",
						LastModified: 0,
					},
				}
			}

			wg.Done()
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

	groupId, err := strconv.ParseUint(vars["id"], 10, 64)

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

	groupPos := config.FindGroup(groupId)

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

type CreateHomePageGroupBody struct {
	Name      string `json:"name"`
	GroupType uint8  `json:"type,omitempty"`
	Prepend   bool   `json:"prepend"`
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

	// Params

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p CreateHomePageGroupBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "The size of the name cannot exceed 255 bytes.")
		return
	}

	if !validateGroupType(p.GroupType) {
		ReturnAPIError(response, 400, "INVALID_GROUP_TYPE", "Invalid group type provided.")
		return
	}

	// Checks

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if len(config.Groups) >= GROUPS_COUNT_LIMIT {
		ReturnAPIError(response, 400, "TOO_MANY_GROUPS", "Your home page exceeds the limit of "+fmt.Sprint(GROUPS_COUNT_LIMIT)+" groups.")
		return
	}

	// Update

	groupId, err := GetVault().homePage.CreateGroup(session.key, p.Name, p.GroupType, p.Prepend)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Result

	res := HomePageGroupApi{
		Id:            groupId,
		Name:          p.Name,
		Type:          p.GroupType,
		ElementsCount: 0,
	}

	jsonResult, err := json.Marshal(res)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type RenameHomePageGroupBody struct {
	Name string `json:"name"`
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

	// Params

	vars := mux.Vars(request)

	groupId, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p RenameHomePageGroupBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Name) > 255 {
		ReturnAPIError(response, 400, "INVALID_NAME", "The size of the name cannot exceed 255 bytes.")
		return
	}

	// Checks

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	groupPos := config.FindGroup(groupId)

	if groupPos == -1 {
		ReturnAPIError(response, 404, "GROUP_NOT_FOUND", "No group was found with the specified ID.")
		return
	}

	// Update

	err = GetVault().homePage.RenameGroup(session.key, groupId, p.Name)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Result

	response.WriteHeader(200)
}

type MoveHomePageGroupBody struct {
	Position int `json:"position"`
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

	// Params

	vars := mux.Vars(request)

	groupId, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p MoveHomePageGroupBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Checks

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	groupPos := config.FindGroup(groupId)

	if groupPos == -1 {
		ReturnAPIError(response, 404, "GROUP_NOT_FOUND", "No group was found with the specified ID.")
		return
	}

	// Update

	err = GetVault().homePage.MoveGroup(session.key, groupId, p.Position)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Result

	response.WriteHeader(200)
}

type HomePageElementMove struct {
	// The element type (media, album)
	ElementType uint8 `json:"t,omitempty"`

	// The ID of the media or the album
	Id uint64 `json:"i"`

	// Position to move the element
	Position int `json:"position,omitempty"`
}

type SetElementsHomePageGroupBody struct {
	Elements      []HomePageElement    `json:"elements,omitempty"`
	AddElement    *HomePageElement     `json:"add,omitempty"`
	DeleteElement *HomePageElement     `json:"delete,omitempty"`
	MoveElement   *HomePageElementMove `json:"move,omitempty"`
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

	// Params

	vars := mux.Vars(request)

	groupId, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p SetElementsHomePageGroupBody

	err = json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Checks

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	groupPos := config.FindGroup(groupId)

	if groupPos == -1 {
		ReturnAPIError(response, 404, "GROUP_NOT_FOUND", "No group was found with the specified ID.")
		return
	}

	if config.Groups[groupPos].Type != HOME_PAGE_GROUP_CUSTOM {
		ReturnAPIError(response, 400, "NOT_CUSTOM_GROUP", "Only custom groups can have elements.")
		return
	}

	if p.AddElement != nil {
		if p.AddElement.ElementType > HOME_PAGE_ELEMENT_TYPE_ALBUM {
			p.AddElement.ElementType = HOME_PAGE_ELEMENT_TYPE_MEDIA
		}

		if len(config.Groups[groupPos].Elements) >= GROUP_ELEMENTS_LIMIT {
			ReturnAPIError(response, 400, "TOO_MANY_ELEMENTS", "Too many elements.")
			return
		}

		err = GetVault().homePage.AddGroupElement(session.key, groupId, *p.AddElement)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	} else if p.DeleteElement != nil {
		err = GetVault().homePage.DeleteGroupElement(session.key, groupId, *p.DeleteElement)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	} else if p.MoveElement != nil {
		err = GetVault().homePage.MoveGroupElement(session.key, groupId, HomePageElement{ElementType: p.MoveElement.ElementType, Id: p.MoveElement.Id}, p.MoveElement.Position)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	} else if p.Elements != nil {
		if len(p.Elements) > GROUP_ELEMENTS_LIMIT {
			ReturnAPIError(response, 400, "TOO_MANY_ELEMENTS", "Too many elements.")
			return
		}

		for i := range p.Elements {
			// Ensure the element types are valid
			if p.Elements[i].ElementType > HOME_PAGE_ELEMENT_TYPE_ALBUM {
				p.Elements[i].ElementType = HOME_PAGE_ELEMENT_TYPE_MEDIA
			}
		}

		err = GetVault().homePage.SetGroupElementList(session.key, groupId, p.Elements)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	// Result

	response.WriteHeader(200)
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

	if !HandleAuthConfirmation(response, request, session, false) {
		return
	}

	// Params

	vars := mux.Vars(request)

	groupId, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Checks

	config, err := GetVault().homePage.Read(session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	groupPos := config.FindGroup(groupId)

	if groupPos == -1 {
		// Already deleted
		response.WriteHeader(200)
		return
	}

	// Update

	err = GetVault().homePage.DeleteGroup(session.key, groupId)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Result

	response.WriteHeader(200)
}
