// Search API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SearchResultResponse struct {
	Count     int64               `json:"total_count"`
	Page      int64               `json:"page_index"`
	PageCount int64               `json:"page_count"`
	PageSize  int64               `json:"page_size"`
	PageItems []*MediaListAPIItem `json:"page_items"`
}

const PAGE_SIZE_LIMIT = 256

func api_searchMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	// Parse arguments

	tagToSearch := ParseTagName(request.URL.Query().Get("tag"))

	reversed := request.URL.Query().Get("order") != "asc"

	var page int64 = 0
	var err error
	pageStr := request.URL.Query().Get("page_index")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)

		if err != nil {
			response.WriteHeader(400)
			return
		}

		if page < 0 {
			page = 0
		}
	}

	var pageSize int64 = 50
	pageSizeStr := request.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		pageSize, err = strconv.ParseInt(pageSizeStr, 10, 64)

		if err != nil || pageSize < 1 {
			response.WriteHeader(400)
			return
		}

		if pageSize > PAGE_SIZE_LIMIT {
			pageSize = PAGE_SIZE_LIMIT
		}
	}

	skip := pageSize * page

	var total_count int64
	var page_items []uint64
	var tag_id uint64

	// Fetch lists from indexes

	if tagToSearch == "" {
		// Default search, use main index
		main_index, err := GetVault().index.StartRead()

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		total_count, err = main_index.Count()

		if err != nil {
			LogError(err)

			GetVault().index.EndRead(main_index)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if reversed {
			page_items, err = main_index.ListValuesReverse(skip, pageSize)

			if err != nil {
				LogError(err)

				GetVault().index.EndRead(main_index)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		} else {
			page_items, err = main_index.ListValues(skip, pageSize)

			if err != nil {
				LogError(err)

				GetVault().index.EndRead(main_index)

				ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
				return
			}
		}

		GetVault().index.EndRead(main_index)
	} else {
		// Search by tag

		page_items, total_count, tag_id, err = GetVault().tags.ListTaggedMedia(tagToSearch, session.key, skip, pageSize, reversed)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	page_count := total_count / pageSize

	if total_count%pageSize > 0 {
		page_count++
	}

	// Read metadata

	page_items_meta := GetMediaMinInfoList(page_items, session)

	for i := 0; i < len(page_items_meta); i++ {
		mediaInfo := page_items_meta[i]

		if mediaInfo.Type == MediaTypeDeleted {
			if tagToSearch != "" {
				err, _ = GetVault().tags.UnTagMedia(page_items[i], tag_id, session.key)

				if err != nil {
					LogError(err)
				}
			} else {
				err = GetVault().index.RemoveElement(page_items[i])

				if err != nil {
					LogError(err)
				}
			}
		}
	}

	// Send response

	var result SearchResultResponse

	result.Count = total_count
	result.Page = page
	result.PageCount = page_count
	result.PageSize = pageSize
	result.PageItems = page_items_meta

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type RandomResultResponse struct {
	Seed      int64               `json:"seed"`
	PageSize  int64               `json:"page_size"`
	PageItems []*MediaListAPIItem `json:"page_items"`
}

func api_randomMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	var err error

	var pageSize int64 = 50
	pageSizeStr := request.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		pageSize, err = strconv.ParseInt(pageSizeStr, 10, 64)

		if err != nil || pageSize < 1 {
			response.WriteHeader(400)
			return
		}

		if pageSize > PAGE_SIZE_LIMIT {
			pageSize = PAGE_SIZE_LIMIT
		}
	}

	var seed int64 = time.Now().UnixMilli()
	seedStr := request.URL.Query().Get("seed")
	if seedStr != "" {
		seed, err = strconv.ParseInt(seedStr, 10, 64)

		if err != nil {
			response.WriteHeader(400)
			return
		}
	}

	tagToSearch := ParseTagName(request.URL.Query().Get("tag"))

	var page_items []uint64
	var tag_id uint64

	if tagToSearch == "" {
		// Default search, use main index
		main_index, err := GetVault().index.StartRead()

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		page_items, err = main_index.RandomValues(seed, pageSize)

		if err != nil {
			LogError(err)

			GetVault().index.EndRead(main_index)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		GetVault().index.EndRead(main_index)
	} else {
		// Search by tag

		page_items, tag_id, err = GetVault().tags.RandomTaggedMedia(tagToSearch, session.key, seed, pageSize)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	// Read meta of media items

	page_items_meta := GetMediaMinInfoList(page_items, session)

	for i := 0; i < len(page_items_meta); i++ {
		mediaInfo := page_items_meta[i]

		if mediaInfo.Type == MediaTypeDeleted {
			if tagToSearch != "" {
				err, _ = GetVault().tags.UnTagMedia(page_items[i], tag_id, session.key)

				if err != nil {
					LogError(err)
				}
			} else {
				err = GetVault().index.RemoveElement(page_items[i])

				if err != nil {
					LogError(err)
				}
			}
		}
	}

	// Send response

	var result RandomResultResponse

	result.Seed = seed
	result.PageSize = pageSize
	result.PageItems = page_items_meta

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type AdvancedSearchResultResponse struct {
	Scanned  int64               `json:"scanned"`
	Count    int64               `json:"total_count"`
	Items    []*MediaListAPIItem `json:"items"`
	Continue uint64              `json:"continue"`
}

func api_advancedSearch(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	// Parameters

	limit := 50
	limitStr := request.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err := strconv.ParseInt(limitStr, 10, 64)

		if err != nil || limit < 1 {
			response.WriteHeader(400)
			return
		}

		if limit > PAGE_SIZE_LIMIT {
			limit = PAGE_SIZE_LIMIT
		}
	}

	tagMode := TagFilterAllOf

	switch strings.ToLower(request.URL.Query().Get("tags_mode")) {
	case "anyof":
		tagMode = TagFilterAnyOf
	case "noneof":
		tagMode = TagFilterNoneOf
	}

	tagNames := make([]string, 0)

	if request.URL.Query().Get("tags") != "" {
		err := json.Unmarshal([]byte(request.URL.Query().Get("tags")), &tagNames)

		if err != nil {
			tagNames = make([]string, 0)
		}
	}

	for i := 0; i < len(tagNames); i++ {
		tagNames[i] = ParseTagName(tagNames[i])
	}

	reversed := request.URL.Query().Get("order") != "asc"

	continueRef := int64(-1)

	if request.URL.Query().Get("continue") != "" {
		cr, err := strconv.ParseInt(request.URL.Query().Get("continue"), 10, 64)

		if err == nil {
			continueRef = cr
		}
	}

	// Create scanner

	LogDebug("[Scanner] TagMode: " + fmt.Sprint(tagMode) + ", Tags: " + fmt.Sprint(tagNames) + ", Continue: " + fmt.Sprint(continueRef))

	vaultScanner, err := NewVaultScanner(tagMode, reversed, tagNames, continueRef, session.key)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if vaultScanner == nil {
		// Empty

		result := AdvancedSearchResultResponse{
			Scanned:  0,
			Count:    0,
			Items:    make([]*MediaListAPIItem, 0),
			Continue: 0,
		}

		jsonResult, err := json.Marshal(result)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		ReturnAPI_JSON(response, request, jsonResult)
		return
	}

	// Scan

	mediaIdList := make([]uint64, 0)
	newContinueRef := uint64(0)

	finished := false

	for !finished && len(mediaIdList) < limit {
		ok, next, err := vaultScanner.Next()

		if err != nil {
			vaultScanner.Release(session.key)

			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		if !ok {
			finished = true
		} else {
			mediaIdList = append(mediaIdList, next)
			newContinueRef = next
		}
	}

	scanned, total := vaultScanner.GetProgress()

	// Release

	vaultScanner.Release(session.key)

	// Read meta of media items

	mediaItemsInfo := GetMediaMinInfoList(mediaIdList, session)

	// Result

	result := AdvancedSearchResultResponse{
		Scanned:  scanned,
		Count:    total,
		Items:    mediaItemsInfo,
		Continue: newContinueRef,
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
