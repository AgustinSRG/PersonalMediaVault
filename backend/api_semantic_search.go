// Semantic search / indexing API

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const MAX_RESULTS_SEMANTIC = 10000

type SearchMediaSemanticBody struct {
	Vector            []float32 `json:"vector"`
	Limit             uint32    `json:"limit"`
	ContinuationToken string    `json:"continuationToken"`
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
			Continue: "",
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
			Continue: "",
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

	if p.Vector == nil || len(p.Vector) != int(semanticSearchStatus.dimensions) {
		ReturnAPIError(response, 400, "INVALID_VECTOR_SIZE", "The size of the vectors are expected to be "+fmt.Sprint(semanticSearchStatus.dimensions))
		return
	}

	if p.Limit > PAGE_SIZE_LIMIT {
		p.Limit = PAGE_SIZE_LIMIT
	} else if p.Limit == 0 {
		p.Limit = 50
	}

	limit := p.Limit

	continuationTokenParts := strings.Split(p.ContinuationToken, ";")
	scanned := int64(0)
	var internalContinuationToken *float32 = nil

	if len(continuationTokenParts) == 2 {
		s, err := strconv.ParseInt(continuationTokenParts[0], 10, 32)

		if err == nil {
			scanned = s
		}

		f, err := strconv.ParseFloat(continuationTokenParts[1], 32)

		if err == nil {
			f2 := float32(f)
			internalContinuationToken = &f2
		}
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

	if total_count > MAX_RESULTS_SEMANTIC {
		total_count = MAX_RESULTS_SEMANTIC
	}

	if uint32(scanned) > uint32(total_count) {
		res := AdvancedSearchResultResponse{
			Scanned:  0,
			Count:    0,
			Items:    make([]*MediaListAPIItem, 0),
			Continue: "",
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

	vectors, nextInternalCt, err := semanticSearch.QueryVectors(request.Context(), &SemanticSearchQuery{
		Vector:            p.Vector,
		Limit:             uint64(limit),
		ContinuationToken: internalContinuationToken,
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

	scanned += int64(len(mediaIdList))

	if scanned > total_count || len(mediaIdList) == 0 || nextInternalCt == nil {
		scanned = total_count
	}

	nextContinuationToken := ""

	if nextInternalCt != nil {
		nextContinuationToken = fmt.Sprint(scanned) + ";" + fmt.Sprint(*nextInternalCt)
	}

	// Read meta of media items

	mediaItemsInfo := GetMediaMinInfoList(mediaIdList, session, func(media_id uint64) {
		semanticSearch.RequestMediaIndexRemoval(media_id, session.key, false)
	})

	// Result

	result := AdvancedSearchResultResponse{
		Scanned:  scanned,
		Count:    total_count,
		Items:    mediaItemsInfo,
		Continue: nextContinuationToken,
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

	vector, err := semanticSearch.ClipEncodeText(request.Context(), p.Text)

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

func api_searchMediaSemanticEncodeImage(response http.ResponseWriter, request *http.Request) {
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

	_, p, err := mime.ParseMediaType(request.Header.Get("Content-Type"))

	if err != nil {
		response.WriteHeader(400)
		return
	}

	boundary := p["boundary"]

	reader := multipart.NewReader(request.Body, boundary)

	part, err := reader.NextPart()

	if err != nil {
		ReturnAPIError(response, 400, "INVALID_MULTIPART_BODY", "Invalid multipart body received")
		return
	}

	mazSize := semanticSearch.GetClipImageSizeLimit()
	image := make([]byte, 0)
	buf := make([]byte, 1024*1024)
	finished := false

	for !finished {
		n, err := part.Read(buf)

		if err != nil && err != io.EOF {
			ReturnAPIError(response, 400, "INVALID_MULTIPART_BODY", "Invalid multipart body received")
			return
		}

		if err == io.EOF {
			finished = true
		}

		if n == 0 {
			continue
		}

		image = append(image, buf[:n]...)

		if int64(len(image)) > mazSize {
			ReturnAPIError(response, 413, "IMAGE_TOO_LARGE", "The image size cannot exceed "+fmt.Sprint(mazSize)+" bytes.")
			return
		}
	}

	// Encode

	vector, isInvalidImageError, err := semanticSearch.ClipEncodeImage(request.Context(), image)

	if isInvalidImageError {
		ReturnAPIError(response, 400, "INVALID_IMAGE", "Invalid image received")
		return
	}

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

func api_searchMediaSemanticGetMediaEmbeddings(response http.ResponseWriter, request *http.Request) {
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

	vars := mux.Vars(request)

	media_id, err := strconv.ParseUint(vars["id"], 10, 64)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	// Get embeddings

	embeddings, err := semanticSearch.GetIndexedVectors(request.Context(), media_id)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if len(embeddings) == 0 {
		ReturnAPIError(response, 404, "EMBEDDINGS_NOT_FOUND", "The specific media ID has no indexed embeddings yet.")
		return
	}

	vectorId := embeddings[0].Id

	vector, err := semanticSearch.GetIndexedVector(request.Context(), vectorId)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	if vector == nil {
		ReturnAPIError(response, 404, "EMBEDDINGS_NOT_FOUND", "The specific media ID has no indexed embeddings yet.")
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
