// Tags API (Bulk)

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type TagAPISetBulkBody struct {
	MediaIds []uint64 `json:"media_ids"`
	TagNames []string `json:"tag_names"`
}

const BULK_TAGS_LIMIT = 64

func api_tagMediaBulk(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	// Validate and parse parameters

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p TagAPISetBulkBody

	err := json.NewDecoder(request.Body).Decode(&p)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.MediaIds) == 0 {
		ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Cannot be empty.")
		return
	}

	if len(p.MediaIds) > BULK_TAGS_LIMIT {
		ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Cannot contain more than "+fmt.Sprint(BULK_TAGS_LIMIT)+" elements.")
		return
	}

	mediaIdSet := make(map[uint64]bool)

	for _, id := range p.MediaIds {
		if mediaIdSet[id] {
			ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Contains duplicated elements.")
			return
		}

		mediaIdSet[id] = true
	}

	if len(p.TagNames) == 0 {
		ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag list: Cannot be empty.")
		return
	}

	if len(p.MediaIds) > BULK_TAGS_LIMIT {
		ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag list: Cannot contain more than "+fmt.Sprint(BULK_TAGS_LIMIT)+" elements.")
		return
	}

	tagNameSet := make(map[string]bool)

	for i, originalTagName := range p.TagNames {
		tagName := ParseTagName(p.TagNames[i])

		if len(tagName) == 0 || len(tagName) > 255 {
			ReturnAPIError(response, 400, "INVALID_TAG_NAME", "Invalid tag name provided: '"+originalTagName+"'")
			return
		}

		p.TagNames[i] = tagName

		if tagNameSet[tagName] {
			ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag list: Contains duplicated elements.")
			return
		}

		tagNameSet[tagName] = true
	}

	// Add media to the indexes in parallel

	wgIndexAdd := &sync.WaitGroup{}

	wgIndexAdd.Add(len(p.TagNames))

	results := make([]*TagListAPIItem, len(p.TagNames))

	for i, tagName := range p.TagNames {
		go applyTagIndexAddBulk(session, wgIndexAdd, tagName, p.MediaIds, results, i)
	}

	wgIndexAdd.Wait()

	tag_ids := make([]uint64, len(results))

	for i, r := range results {
		if r == nil {
			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}

		tag_ids[i] = r.Id
	}

	// Add tags to the metadata of the media

	wgMetadataUpdate := &sync.WaitGroup{}

	wgMetadataUpdate.Add(len(p.MediaIds))

	successMarks := make([]bool, len(p.MediaIds))

	for i, media_id := range p.MediaIds {
		go applyTagMediaMetadataAddBulk(session, wgMetadataUpdate, media_id, tag_ids, successMarks, i)
	}

	wgMetadataUpdate.Wait()

	for _, success := range successMarks {
		if !success {
			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	// Send result

	jsonResult, err := json.Marshal(results)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func applyTagIndexAddBulk(session *ActiveSession, wg *sync.WaitGroup, tagName string, media_ids []uint64, results []*TagListAPIItem, resultIndex int) {
	defer wg.Done()

	tag_id, err := GetVault().tags.TagMediaBulk(media_ids, tagName, session.key)

	if err != nil {
		LogError(err)
		results[resultIndex] = nil
		return
	}

	results[resultIndex] = &TagListAPIItem{
		Id:   tag_id,
		Name: tagName,
	}
}

func applyTagMediaMetadataAddBulk(session *ActiveSession, wg *sync.WaitGroup, media_id uint64, tag_ids []uint64, successMarks []bool, markIndex int) {
	defer wg.Done()

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		// Media not found, skip
		successMarks[markIndex] = true
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		successMarks[markIndex] = false
		return
	}

	if meta == nil {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)

		// Media not found, skip

		successMarks[markIndex] = true
		return
	}

	for _, tag_id := range tag_ids {
		meta.AddTag(tag_id)
	}

	err = media.EndWrite(meta, session.key, false)

	GetVault().media.ReleaseMediaResource(media_id)

	// Clear cache

	GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(media_id)

	if err != nil {
		LogError(err)

		successMarks[markIndex] = false
		return
	}

	// Mark success

	successMarks[markIndex] = true
}

type TagAPIRemoveBulkBody struct {
	MediaIds []uint64 `json:"media_ids"`
	TagIds   []uint64 `json:"tag_ids"`
}

type UntagMediaBulkResponse struct {
	Removed []bool `json:"removed"`
}

func api_untagMediaBulk(response http.ResponseWriter, request *http.Request) {
	session := GetSessionFromRequest(request)

	if session == nil {
		ReturnAPIError(response, 401, "UNAUTHORIZED", "You must provide a valid active session to use this API.")
		return
	}

	if !session.CanWrite() {
		ReturnAPIError(response, 403, "ACCESS_DENIED", "Your current session does not have permission to make use of this API.")
		return
	}

	// Validate and parse parameters

	request.Body = http.MaxBytesReader(response, request.Body, JSON_BODY_MAX_LENGTH)

	var p TagAPIRemoveBulkBody

	err := json.NewDecoder(request.Body).Decode(&p)

	if err != nil {
		response.WriteHeader(400)
		return
	}

	if len(p.MediaIds) == 0 {
		ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Cannot be empty.")
		return
	}

	if len(p.MediaIds) > BULK_TAGS_LIMIT {
		ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Cannot contain more than "+fmt.Sprint(BULK_TAGS_LIMIT)+" elements.")
		return
	}

	mediaIdSet := make(map[uint64]bool)

	for _, id := range p.MediaIds {
		if mediaIdSet[id] {
			ReturnAPIError(response, 400, "INVALID_MEDIA_LIST", "Invalid media ID list: Contains duplicated elements.")
			return
		}

		mediaIdSet[id] = true
	}

	if len(p.TagIds) == 0 {
		ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag ID list: Cannot be empty.")
		return
	}

	if len(p.TagIds) > BULK_TAGS_LIMIT {
		ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag ID list: Cannot contain more than "+fmt.Sprint(BULK_TAGS_LIMIT)+" elements.")
		return
	}

	tagIdSet := make(map[uint64]bool)

	for _, id := range p.TagIds {
		if tagIdSet[id] {
			ReturnAPIError(response, 400, "INVALID_TAG_LIST", "Invalid tag ID list: Contains duplicated elements.")
			return
		}

		tagIdSet[id] = true
	}

	// Remove media from the indexes in parallel

	wgIndexRemove := &sync.WaitGroup{}

	wgIndexRemove.Add(len(p.TagIds))

	successMarks := make([]bool, len(p.TagIds))
	removedMarks := make([]bool, len(p.TagIds))

	for i, tagId := range p.TagIds {
		go applyTagIndexRemoveBulk(session, wgIndexRemove, tagId, p.MediaIds, successMarks, removedMarks, i)
	}

	wgIndexRemove.Wait()

	for _, success := range successMarks {
		if !success {
			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	// Remove tags from the metadata of the media

	wgMetadataUpdate := &sync.WaitGroup{}

	wgMetadataUpdate.Add(len(p.MediaIds))

	successMarks = make([]bool, len(p.MediaIds))

	for i, media_id := range p.MediaIds {
		go applyTagMediaMetadataRemoveBulk(session, wgMetadataUpdate, media_id, p.TagIds, successMarks, i)
	}

	wgMetadataUpdate.Wait()

	for _, success := range successMarks {
		if !success {
			ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
			return
		}
	}

	// Send result

	var result UntagMediaBulkResponse

	result.Removed = removedMarks

	jsonResult, err := json.Marshal(result)

	if err != nil {
		LogError(err)

		ReturnAPIError(response, 500, "INTERNAL_ERROR", "Internal server error, Check the logs for details.")
		return
	}

	ReturnAPI_JSON(response, request, jsonResult)
}

func applyTagIndexRemoveBulk(session *ActiveSession, wg *sync.WaitGroup, tagId uint64, media_ids []uint64, successMarks []bool, removedMarks []bool, bulkIndex int) {
	defer wg.Done()

	err, tagWasRemoved := GetVault().tags.UnTagMediaBulk(media_ids, tagId, session.key)

	if err != nil {
		LogError(err)

		successMarks[bulkIndex] = false
		return
	}

	removedMarks[bulkIndex] = tagWasRemoved
	successMarks[bulkIndex] = true
}

func applyTagMediaMetadataRemoveBulk(session *ActiveSession, wg *sync.WaitGroup, media_id uint64, tag_ids []uint64, successMarks []bool, markIndex int) {
	defer wg.Done()

	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		// Media not found, skip
		successMarks[markIndex] = true
		return
	}

	meta, err := media.StartWrite(session.key)

	if err != nil {
		LogError(err)

		GetVault().media.ReleaseMediaResource(media_id)

		successMarks[markIndex] = false
		return
	}

	if meta == nil {
		media.CancelWrite()
		GetVault().media.ReleaseMediaResource(media_id)

		// Media not found, skip

		successMarks[markIndex] = true
		return
	}

	for _, tag_id := range tag_ids {
		meta.RemoveTag(tag_id)
	}

	err = media.EndWrite(meta, session.key, false)

	GetVault().media.ReleaseMediaResource(media_id)

	// Clear cache

	GetVault().media.preview_cache.RemoveEntryOrMarkInvalid(media_id)

	if err != nil {
		LogError(err)

		successMarks[markIndex] = false
		return
	}

	// Mark success

	successMarks[markIndex] = true
}
