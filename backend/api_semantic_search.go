// Semantic search / indexing API

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const MAX_RESULTS_SEMANTIC = 10000

type SearchMediaSemanticBody struct {
	Vector            []float32 `json:"vector"`
	VectorType        string    `json:"vectorType"`
	Limit             uint32    `json:"limit"`
	ContinuationToken uint32    `json:"continuationToken"`
}

func api_searchMediaSemantic(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	// Check if system is available

	semanticSearch := GetVault().semanticSearch

	if semanticSearch == nil {
		res := AdvancedSearchResultResponse{
			Scanned:  0,
			Count:    0,
			Items:    make([]*MediaListAPIItem, 0),
			Continue: 0,
		}

		jsonResult, err := json.Marshal(res)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		ReturnAPI_JSON(response, request, jsonResult)
	}

	semanticSearchStatus := semanticSearch.GetStatus()

	if !semanticSearchStatus.available {
		res := AdvancedSearchResultResponse{
			Scanned:  0,
			Count:    0,
			Items:    make([]*MediaListAPIItem, 0),
			Continue: 0,
		}

		jsonResult, err := json.Marshal(res)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		ReturnAPI_JSON(response, request, jsonResult)
	}

	// Params

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p SearchMediaSemanticBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if p.Vector == nil || len(p.Vector) != int(semanticSearchStatus.clipModelDimensions) {
		ReturnAPIError(response, 400, "INVALID_VECTOR_SIZE", "The size of the vectors are expected to be "+fmt.Sprint(semanticSearchStatus.clipModelDimensions))
		return
	}

	if p.Limit > PAGE_SIZE_LIMIT {
		p.Limit = PAGE_SIZE_LIMIT
	} else if p.Limit == 0 {
		p.Limit = 50
	}

	limit := p.Limit

	var vectorType *QdrantIndexedVectorType = nil

	switch p.VectorType {
	case "text":
		vectorTypeVal := VECTOR_TYPE_TEXT
		vectorType = &vectorTypeVal
	case "image":
		vectorTypeVal := VECTOR_TYPE_IMAGE
		vectorType = &vectorTypeVal
	}

	// Count

	main_index, err := GetVault().index.StartRead()

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	total_count, err := main_index.Count()

	if err != nil {
		LogError(err)

		GetVault().index.EndRead(main_index)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	GetVault().index.EndRead(main_index)

	// Calculate offset

	if vectorType == nil {
		total_count = total_count * 2
	}

	if total_count > MAX_RESULTS_SEMANTIC {
		total_count = MAX_RESULTS_SEMANTIC
	}

	if p.ContinuationToken > uint32(total_count) {
		res := AdvancedSearchResultResponse{
			Scanned:  0,
			Count:    0,
			Items:    make([]*MediaListAPIItem, 0),
			Continue: 0,
		}

		jsonResult, err := json.Marshal(res)

		if err != nil {
			LogError(err)

			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		ReturnAPI_JSON(response, request, jsonResult)
	}

	// Find results

	vectors, err := semanticSearch.QueryVectors(request.Context(), &SemanticSearchQuery{
		Vector:     p.Vector,
		VectorType: vectorType,
		Limit:      uint64(limit),
		Offset:     uint64(p.ContinuationToken),
	})

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	mediaIdList := make([]uint64, len(vectors))

	for i := range vectors {
		mediaIdList[i] = vectors[i].Media
	}

	scanned := int64(p.ContinuationToken) + int64(len(mediaIdList))

	if scanned > total_count || len(mediaIdList) == 0 {
		scanned = total_count
	}

	// Read meta of media items

	mediaItemsInfo := GetMediaMinInfoList(mediaIdList, session)

	// Result

	result := AdvancedSearchResultResponse{
		Scanned:  scanned,
		Count:    total_count,
		Items:    mediaItemsInfo,
		Continue: uint64(scanned),
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

type SearchMediaSemanticEncodeTextBody struct {
	Text string `json:"text"`
}

type SearchMediaSemanticEncodeResponse struct {
	Vector []float32 `json:"vector"`
}

const MAX_TEXT_ENCODE_SIZE = 300

func api_searchMediaSemanticEncodeText(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	// Check if system is available

	semanticSearch := GetVault().semanticSearch

	if semanticSearch == nil {
		ReturnAPIError(response, 404, "SYSTEM_UNAVAILABLE", "The semantic search sub-system is unavailable.")
		return
	}

	semanticSearchStatus := semanticSearch.GetStatus()

	if !semanticSearchStatus.available {
		ReturnAPIError(response, 404, "SYSTEM_UNAVAILABLE", "The semantic search sub-system is unavailable.")
		return
	}

	// Params

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p SearchMediaSemanticEncodeTextBody

	err := json.NewDecoder(request.Body).Decode(&p)
	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.Text) == 0 {
		ReturnAPIError(response, 400, "EMPTY_TEXT", "The text cannot be empty")
		return
	}

	if len(p.Text) > MAX_TEXT_ENCODE_SIZE {
		p.Text = p.Text[0:MAX_TEXT_ENCODE_SIZE]
	}

	// Encode

	vector, err := semanticSearch.ClipEncodeText(p.Text)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	// Response

	result := SearchMediaSemanticEncodeResponse{
		Vector: vector,
	}

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}
