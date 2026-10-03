// API Test

package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func Tags_API_Test_TagMedia(server *httptest.Server, session string, t *testing.T, mediaId uint64, tag string) (tagId uint64, e error) {
	body, err := json.Marshal(TagAPISetBody{
		Media: mediaId,
		Tag:   tag,
	})

	if err != nil {
		t.Error(err)
		return 0, err
	}

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "POST", "/api/tags/add", body, session)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res := TagListAPIItem{}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	if res.Name != tag {
		t.Error(ErrorMismatch("TagName", fmt.Sprint(res.Name), tag))
	}

	meta := MediaAPIMetaResponse{}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return 0, err
	}

	containsTag := false

	for i := 0; i < len(meta.Tags); i++ {
		if meta.Tags[i] == res.Id {
			containsTag = true
			break
		}
	}

	if !containsTag {
		t.Errorf("Media does not contain the tag")
	}

	return res.Id, nil
}

func Tags_API_Test_UntagMedia(server *httptest.Server, session string, t *testing.T, mediaId uint64, tag uint64) {
	body, err := json.Marshal(UntagMediaBody{
		Media: mediaId,
		Tag:   tag,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err := DoTestRequest(server, "POST", "/api/tags/remove", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	meta := MediaAPIMetaResponse{}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag := false

	for i := 0; i < len(meta.Tags); i++ {
		if meta.Tags[i] == tag {
			containsTag = true
			break
		}
	}

	if containsTag {
		t.Errorf("Media contains the tag")
	}
}

func Tags_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Upload test media
	media1, _, err := UploadTestMedia(server, session, MediaTypeImage, "Test Tagged 1", "")

	if err != nil {
		t.Error(err)
		return
	}

	// Tag media

	testTag, err := Tags_API_Test_TagMedia(server, session, t, media1, "test_tag")

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/tags", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res := make([]TagListAPIItem, 0)

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag := false

	for i := 0; i < len(res); i++ {
		if res[i].Id == testTag {
			containsTag = true
			if res[i].Name != "test_tag" {
				t.Error(ErrorMismatch("TagName", res[i].Name, "test_tag"))
			}
			break
		}
	}

	if !containsTag {
		t.Errorf("Tag list does not contain the tag")
	}

	// Untag media

	Tags_API_Test_UntagMedia(server, session, t, media1, testTag)

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/tags", nil, session)

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

	containsTag = false

	for i := 0; i < len(res); i++ {
		if res[i].Id == testTag {
			containsTag = true
			break
		}
	}

	if containsTag {
		t.Errorf("Tag list contains the deleted tag")
	}
}

func Tags_API_Bulk_Test(server *httptest.Server, session string, t *testing.T) {
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

	// Test tags

	testTag1Name := "bulk_tag_1"
	testTag2Name := "bulk_tag_2"
	testTag3Name := "bulk_tag_3"

	// Add tags (bulk)

	body, err := json.Marshal(TagAPISetBulkBody{
		MediaIds: []uint64{media1, media2},
		TagNames: []string{testTag1Name, testTag2Name, testTag3Name},
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "POST", "/api/tags/add_bulk", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		t.Error("Response: " + string(bodyResponseBytes))
	}

	res := []TagListAPIItem{}

	err = json.Unmarshal(bodyResponseBytes, &res)

	if err != nil {
		t.Error(err)
		return
	}

	if len(res) != 3 || res[0].Name != testTag1Name || res[1].Name != testTag2Name || res[2].Name != testTag3Name {
		t.Errorf("Invalid tag add response: %v", string(bodyResponseBytes))
		return
	}

	tag1 := res[0].Id
	tag2 := res[1].Id
	tag3 := res[2].Id

	meta := MediaAPIMetaResponse{}

	// Check Media 1

	err = _TestFetchMetadata(server, session, t, media1, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag1 := false
	containsTag2 := false
	containsTag3 := false

	for i := 0; i < len(meta.Tags); i++ {
		switch meta.Tags[i] {
		case tag1:
			containsTag1 = true
		case tag2:
			containsTag2 = true
		case tag3:
			containsTag3 = true
		}
	}

	if !containsTag1 {
		t.Error("Media1's tag list does not contain " + testTag1Name)
	}

	if !containsTag2 {
		t.Error("Media1's tag list does not contain " + testTag2Name)
	}

	if !containsTag3 {
		t.Error("Media1's tag list does not contain " + testTag3Name)
	}

	// Check Media 2

	err = _TestFetchMetadata(server, session, t, media2, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag1 = false
	containsTag2 = false
	containsTag3 = false

	for i := 0; i < len(meta.Tags); i++ {
		switch meta.Tags[i] {
		case tag1:
			containsTag1 = true
		case tag2:
			containsTag2 = true
		case tag3:
			containsTag3 = true
		}
	}

	if !containsTag1 {
		t.Error("Media2's tag list does not contain " + testTag1Name)
	}

	if !containsTag2 {
		t.Error("Media2's tag list does not contain " + testTag2Name)
	}

	if !containsTag3 {
		t.Error("Media2's tag list does not contain " + testTag3Name)
	}

	// Remove tags (bulk)

	body, err = json.Marshal(TagAPIRemoveBulkBody{
		MediaIds: []uint64{media1, media2},
		TagIds:   []uint64{tag1, tag3},
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "POST", "/api/tags/remove_bulk", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res2 := UntagMediaBulkResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	if len(res2.Removed) != 2 || !res2.Removed[0] || !res2.Removed[1] {
		t.Errorf("Invalid tag remove bulk response: %v", string(bodyResponseBytes))
		return
	}

	// Check Media 1

	err = _TestFetchMetadata(server, session, t, media1, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag1 = false
	containsTag2 = false
	containsTag3 = false

	for i := 0; i < len(meta.Tags); i++ {
		switch meta.Tags[i] {
		case tag1:
			containsTag1 = true
		case tag2:
			containsTag2 = true
		case tag3:
			containsTag3 = true
		}
	}

	if containsTag1 {
		t.Error("Media1's tag list contains " + testTag1Name)
	}

	if !containsTag2 {
		t.Error("Media1's tag list does not contain " + testTag2Name)
	}

	if containsTag3 {
		t.Error("Media1's tag list contains " + testTag3Name)
	}

	// Check Media 2

	err = _TestFetchMetadata(server, session, t, media2, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	containsTag1 = false
	containsTag2 = false
	containsTag3 = false

	for i := 0; i < len(meta.Tags); i++ {
		switch meta.Tags[i] {
		case tag1:
			containsTag1 = true
		case tag2:
			containsTag2 = true
		case tag3:
			containsTag3 = true
		}
	}

	if containsTag1 {
		t.Error("Media2's tag list contains " + testTag1Name)
	}

	if !containsTag2 {
		t.Error("Media2's tag list does not contain " + testTag2Name)
	}

	if containsTag3 {
		t.Error("Media2's tag list contains " + testTag3Name)
	}
}
