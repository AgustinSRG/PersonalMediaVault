// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
)

func Search_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Upload test media
	media1, _, err := UploadTestMedia(server, session, MediaTypeImage, "Test Tagged 1", "")

	if err != nil {
		t.Error(err)
		return
	}

	media2, _, err := UploadTestMedia(server, session, MediaTypeImage, "Test Tagged 2", "")

	if err != nil {
		t.Error(err)
		return
	}

	media3, _, err := UploadTestMedia(server, session, MediaTypeImage, "Test Tagged 3", "")

	if err != nil {
		t.Error(err)
		return
	}

	// Search all

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/search", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res := SearchResultResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	containsMedia := make(map[uint64]bool)

	for i := 0; i < len(res.PageItems); i++ {
		containsMedia[res.PageItems[i].Id] = true
	}

	if !containsMedia[media1] {
		t.Errorf("Search results does not contain media 1")
	}

	if !containsMedia[media2] {
		t.Errorf("Search results does not contain media 2")
	}

	if !containsMedia[media3] {
		t.Errorf("Search results does not contain media 3")
	}

	// Search by tag

	searchTag, err := Tags_API_Test_TagMedia(server, session, t, media1, "search_tag")

	if err != nil {
		t.Error(err)
		return
	}

	_, err = Tags_API_Test_TagMedia(server, session, t, media3, "search_tag")

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/search?tag="+url.QueryEscape("search_tag"), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	containsMedia = make(map[uint64]bool)

	for i := 0; i < len(res.PageItems); i++ {
		containsMedia[res.PageItems[i].Id] = true
	}

	if !containsMedia[media1] {
		t.Errorf("Search results does not contain media 1")
	}

	if containsMedia[media2] {
		t.Errorf("Search results contains media 2")
	}

	if !containsMedia[media3] {
		t.Errorf("Search results does not contain media 3")
	}

	// Untag and search

	Tags_API_Test_UntagMedia(server, session, t, media1, searchTag)

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/search?tag="+url.QueryEscape("search_tag"), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	containsMedia = make(map[uint64]bool)

	for i := 0; i < len(res.PageItems); i++ {
		containsMedia[res.PageItems[i].Id] = true
	}

	if containsMedia[media1] {
		t.Errorf("Search results contains media 1")
	}

	if containsMedia[media2] {
		t.Errorf("Search results contains media 2")
	}

	if !containsMedia[media3] {
		t.Errorf("Search results does not contain media 3")
	}

	// Random

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/random", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res2 := RandomResultResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	// Random filtered by tag

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/random?tag="+url.QueryEscape("search_tag"), nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < len(res2.PageItems); i++ {
		hasTag := false

		for j := 0; j < len(res2.PageItems[i].Tags); j++ {
			if res2.PageItems[i].Tags[j] == searchTag {
				hasTag = true
				break
			}
		}

		if !hasTag {
			t.Errorf("Random media does not contain the filter tag")
		}
	}
}
