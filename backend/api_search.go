// Search API

package main

import (
	"encoding/json"
	"net/http"
	"strconv"
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
		response.WriteHeader(401)
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

	// Fetch lists from indexes

	if tagToSearch == "" {
		// Default search, use main index
		main_index, err := GetVault().index.StartRead()

		if err != nil {
			LogError(err)

			response.WriteHeader(500)
			return
		}

		total_count, err = main_index.Count()

		if err != nil {
			LogError(err)

			GetVault().index.EndRead(main_index)

			response.WriteHeader(500)
			return
		}

		if reversed {
			page_items, err = main_index.ListValuesReverse(skip, pageSize)

			if err != nil {
				LogError(err)

				GetVault().index.EndRead(main_index)

				response.WriteHeader(500)
				return
			}
		} else {
			page_items, err = main_index.ListValues(skip, pageSize)

			if err != nil {
				LogError(err)

				GetVault().index.EndRead(main_index)

				response.WriteHeader(500)
				return
			}
		}

		GetVault().index.EndRead(main_index)
	} else {
		// Search by tag

		page_items, total_count, err = GetVault().tags.ListTaggedMedia(tagToSearch, session.key, skip, pageSize, reversed)

		if err != nil {
			LogError(err)

			response.WriteHeader(500)
			return
		}
	}

	page_count := total_count / pageSize

	if total_count%pageSize > 0 {
		page_count++
	}

	// Read metadata

	page_items_meta := make([]*MediaListAPIItem, len(page_items))

	for i := 0; i < len(page_items); i++ {
		page_items_meta[i] = GetMediaMinInfo(page_items[i], session)
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

		response.WriteHeader(500)
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.Header().Add("Cache-Control", "no-cache")
	response.WriteHeader(200)

	response.Write(jsonResult)
}

type RandomResultResponse struct {
	Seed      int64               `json:"seed"`
	PageSize  int64               `json:"page_size"`
	PageItems []*MediaListAPIItem `json:"page_items"`
}

func api_randomMedia(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		response.WriteHeader(401)
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

	if tagToSearch == "" {
		// Default search, use main index
		main_index, err := GetVault().index.StartRead()

		if err != nil {
			LogError(err)

			response.WriteHeader(500)
			return
		}

		page_items, err = main_index.RandomValues(seed, pageSize)

		if err != nil {
			LogError(err)

			GetVault().index.EndRead(main_index)

			response.WriteHeader(500)
			return
		}

		GetVault().index.EndRead(main_index)
	} else {
		// Search by tag

		page_items, err = GetVault().tags.RandomTaggedMedia(tagToSearch, session.key, seed, pageSize)

		if err != nil {
			LogError(err)

			response.WriteHeader(500)
			return
		}
	}

	// Read meta of media items

	page_items_meta := make([]*MediaListAPIItem, len(page_items))

	for i := 0; i < len(page_items); i++ {
		page_items_meta[i] = GetMediaMinInfo(page_items[i], session)
	}

	// Send response

	var result RandomResultResponse

	result.Seed = seed
	result.PageSize = pageSize
	result.PageItems = page_items_meta

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
