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
