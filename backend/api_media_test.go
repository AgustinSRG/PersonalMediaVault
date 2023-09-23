// API Test

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func Media_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Upload media and wait for it to be ready

	t.Run("Video", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeVideo, "Test video", "")

		if err != nil {
			t.Error(err)
			return
		}

		_TestUploadedMedia(server, session, t, mid)
	})

	t.Run("Audio", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeAudio, "Test audio", "")

		if err != nil {
			t.Error(err)
			return
		}

		_TestUploadedMedia(server, session, t, mid)
	})

	t.Run("Image", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeImage, "Test image", "")

		if err != nil {
			t.Error(err)
			return
		}

		_TestUploadedMedia(server, session, t, mid)
	})
}

func _TestFetchMetadata(server *httptest.Server, session string, t *testing.T, mediaId uint64, res *MediaAPIMetaResponse) error {
	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId)), nil, session)

	if err != nil {
		return err
	}

	if statusCode != 200 {
		return ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200")
	}

	return json.Unmarshal(bodyResponseBytes, &res)
}

func _TestUploadedMedia(server *httptest.Server, session string, t *testing.T, mediaId uint64) {
	// Get metadata

	meta := MediaAPIMetaResponse{}

	err := _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	// Test size stats

	statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/size_stats", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	res1 := MediaSizeAPIResponse{}

	err = json.Unmarshal(bodyResponseBytes, &res1)

	if err != nil {
		t.Error(err)
		return
	}

	foundOriginal := false

	for i := 0; i < len(res1.AssetSize); i++ {
		if res1.AssetSize[i].Name == "ORIGINAL" {
			foundOriginal = true
			break
		}
	}

	if !foundOriginal {
		t.Errorf("Original asset not found in the size stats list")
	}

	// Get original

	statusCode, _, err = DoTestRequest(server, "GET", meta.Url, nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	// Get thumbnail (if any)

	if meta.Thumbnail != "" {
		statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", meta.Thumbnail, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		thumbContentType := http.DetectContentType(bodyResponseBytes)

		if thumbContentType != "image/jpeg" {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(thumbContentType), "image/jpeg"))
		}
	}

	// Change title

	newTitle := "New test media title - " + fmt.Sprint(mediaId)

	body, err := json.Marshal(MediaAPIEditTitleBody{
		Title: newTitle,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/title", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	if meta.Title != newTitle {
		t.Error(ErrorMismatch("Title", meta.Title, newTitle))
	}

	// Change description

	newDescription := "New test media description - " + fmt.Sprint(mediaId)

	body, err = json.Marshal(MediaAPIEditDescriptionBody{
		Description: newDescription,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/description", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	if meta.Description != newDescription {
		t.Error(ErrorMismatch("Description", meta.Title, newTitle))
	}

	// Change extra data

	body, err = json.Marshal(MediaAPIEditExtraParams{
		ForceStartBeginning: true,
	})

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/extra", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	if !meta.ForceStartBeginning {
		t.Error(ErrorMismatch("ForceStartBeginning", fmt.Sprint(meta.ForceStartBeginning), fmt.Sprint(true)))
	}

	// Time slices

	if meta.Type == MediaTypeAudio || meta.Type == MediaTypeVideo {
		newTimeSlices := []MediaAPIMetaTimeSplit{
			{Time: 0, Name: "Slice 1"},
			{Time: 5, Name: "Slice 2"},
		}

		body, err = json.Marshal(newTimeSlices)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/time_slices", body, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(meta.TimeSlices, newTimeSlices) {
			t.Error(ErrorMismatch("TimeSlices", fmt.Sprint(meta.TimeSlices), fmt.Sprint(newTimeSlices)))
		}
	}

	// Image notes

	if meta.Type == MediaTypeImage {
		newNotes := []ImageNote{
			{
				XPosition: 0,
				YPosition: 0,
				Width:     100,
				Height:    100,
				Text:      "Note 1",
			},
			{
				XPosition: 7,
				YPosition: 7,
				Width:     777,
				Height:    777,
				Text:      "Note 2",
			},
		}

		body, err = json.Marshal(newNotes)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/notes", body, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if !meta.HasImageNotes {
			t.Error(ErrorMismatch("HasImageNotes", fmt.Sprint(meta.HasImageNotes), fmt.Sprint(true)))
		}

		if meta.ImageNotesURL != "" {
			statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", meta.ImageNotesURL, nil, session)

			if err != nil {
				t.Error(err)
				return
			}

			if statusCode != 200 {
				t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
			}

			metaNotes := make([]ImageNote, 0)

			err = json.Unmarshal(bodyResponseBytes, &metaNotes)

			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(metaNotes, newNotes) {
				t.Error(ErrorMismatch("Notes", fmt.Sprint(metaNotes), fmt.Sprint(newNotes)))
			}
		} else {
			t.Errorf("ImageNotesURL is empty")
		}
	}

	// Extended description

	newExtDesc := ExtendedDescriptionSetBody{
		ExtendedDescription: "Test extended description",
	}

	body, err = json.Marshal(newExtDesc)

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/ext_desc", body, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	if meta.ExtendedDescriptionURL != "" {
		statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", meta.ExtendedDescriptionURL, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		bodyResponseString := string(bodyResponseBytes)

		if bodyResponseString != newExtDesc.ExtendedDescription {
			t.Error(ErrorMismatch("ExtendedDescription", bodyResponseString, newExtDesc.ExtendedDescription))
		}
	} else {
		t.Errorf("ExtendedDescriptionURL is empty")
	}

	// Change thumbnail

	client := server.Client()

	newThumbnail, err := os.ReadFile(path.Join("test-assets", "test-image.png"))

	if err != nil {
		t.Error(err)
		return
	}

	var b bytes.Buffer

	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", "new-thumbnail.png")

	if err != nil {
		t.Error(err)
		return
	}

	_, err = part.Write(newThumbnail)

	if err != nil {
		t.Error(err)
		return
	}

	writer.Close()

	apiURL, err := url.JoinPath(server.URL, "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/edit/thumbnail")

	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", apiURL, &b)

	if err != nil {
		t.Error(err)
		return
	}

	req.Header.Set("x-session-token", session)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)

	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	bodyData, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Error(err)
		return
	}

	res2 := ThumbnailAPIResponse{}

	err = json.Unmarshal(bodyData, &res2)

	if err != nil {
		t.Error(err)
		return
	}

	statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", res2.Url, nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	thumbContentType := http.DetectContentType(bodyResponseBytes)

	if thumbContentType != "image/jpeg" {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(thumbContentType), "image/jpeg"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	if meta.Thumbnail == "" {
		t.Errorf("Metadata has no thumbnail")
	}

	// Subtitles

	if meta.Type == MediaTypeVideo || meta.Type == MediaTypeAudio {
		newSubtitles, err := os.ReadFile(path.Join("test-assets", "test-subrip.srt"))

		if err != nil {
			t.Error(err)
			return
		}

		var b bytes.Buffer

		writer := multipart.NewWriter(&b)

		part, err := writer.CreateFormFile("file", "test-subrip.srt")

		if err != nil {
			t.Error(err)
			return
		}

		_, err = part.Write(newSubtitles)

		if err != nil {
			t.Error(err)
			return
		}

		writer.Close()

		apiURL, err := url.JoinPath(server.URL, "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/subtitles/set")

		if err != nil {
			t.Error(err)
			return
		}

		req, err := http.NewRequest("POST", apiURL+"?id=en&name=English", &b)

		if err != nil {
			t.Error(err)
			return
		}

		req.Header.Set("x-session-token", session)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)

		if err != nil {
			t.Error(err)
			return
		}

		defer resp.Body.Close()

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		bodyData, err := io.ReadAll(resp.Body)

		if err != nil {
			t.Error(err)
			return
		}

		res2 := SubtitlesAPIResponse{}

		err = json.Unmarshal(bodyData, &res2)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", res2.Url, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		if !reflect.DeepEqual(newSubtitles, bodyResponseBytes) {
			t.Errorf("Subtitles content mismatch")
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Subtitles) != 1 {
			t.Error(ErrorMismatch("len(Subtitles)", fmt.Sprint(len(meta.Subtitles)), fmt.Sprint(1)))
		}

		if meta.Subtitles[0].Id != "en" || meta.Subtitles[0].Name != "English" {
			t.Error(ErrorMismatch("len(Subtitles)", meta.Subtitles[0].Id+"/"+meta.Subtitles[0].Name, "en/English"))
		}

		// Delete the subtitles

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/subtitles/remove?id=en", nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Subtitles) != 0 {
			t.Error(ErrorMismatch("len(Subtitles)", fmt.Sprint(len(meta.Subtitles)), fmt.Sprint(0)))
		}
	}

	// Audio tracks

	if meta.Type == MediaTypeVideo {
		newAudio, err := os.ReadFile(path.Join("test-assets", "test-audio.mp3"))

		if err != nil {
			t.Error(err)
			return
		}

		var b bytes.Buffer

		writer := multipart.NewWriter(&b)

		part, err := writer.CreateFormFile("file", "audio.mp3")

		if err != nil {
			t.Error(err)
			return
		}

		_, err = part.Write(newAudio)

		if err != nil {
			t.Error(err)
			return
		}

		writer.Close()

		apiURL, err := url.JoinPath(server.URL, "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/audios/set")

		if err != nil {
			t.Error(err)
			return
		}

		req, err := http.NewRequest("POST", apiURL+"?id=en&name=English", &b)

		if err != nil {
			t.Error(err)
			return
		}

		req.Header.Set("x-session-token", session)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)

		if err != nil {
			t.Error(err)
			return
		}

		defer resp.Body.Close()

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		bodyData, err := io.ReadAll(resp.Body)

		if err != nil {
			t.Error(err)
			return
		}

		res2 := AudioTrackAPIResponse{}

		err = json.Unmarshal(bodyData, &res2)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", res2.Url, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		if !reflect.DeepEqual(newAudio, bodyResponseBytes) {
			t.Errorf("Audios content mismatch")
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Audios) != 1 {
			t.Error(ErrorMismatch("len(Audios)", fmt.Sprint(len(meta.Audios)), fmt.Sprint(1)))
		}

		if meta.Audios[0].Id != "en" || meta.Audios[0].Name != "English" {
			t.Error(ErrorMismatch("Audios", meta.Audios[0].Id+"/"+meta.Audios[0].Name, "en/English"))
		}

		// Delete the audio

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/audios/remove?id=en", nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Audios) != 0 {
			t.Error(ErrorMismatch("len(Audios)", fmt.Sprint(len(meta.Audios)), fmt.Sprint(0)))
		}
	}

	// Attachments

	if meta.Type == MediaTypeVideo || meta.Type == MediaTypeAudio || meta.Type == MediaTypeImage {
		newAttachment, err := os.ReadFile(path.Join("test-assets", "test-attachment.txt"))

		if err != nil {
			t.Error(err)
			return
		}

		var b bytes.Buffer

		writer := multipart.NewWriter(&b)

		part, err := writer.CreateFormFile("file", "test-attachment.txt")

		if err != nil {
			t.Error(err)
			return
		}

		_, err = part.Write(newAttachment)

		if err != nil {
			t.Error(err)
			return
		}

		writer.Close()

		apiURL, err := url.JoinPath(server.URL, "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/attachments/add")

		if err != nil {
			t.Error(err)
			return
		}

		req, err := http.NewRequest("POST", apiURL, &b)

		if err != nil {
			t.Error(err)
			return
		}

		req.Header.Set("x-session-token", session)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := client.Do(req)

		if err != nil {
			t.Error(err)
			return
		}

		defer resp.Body.Close()

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		bodyData, err := io.ReadAll(resp.Body)

		if err != nil {
			t.Error(err)
			return
		}

		res2 := AttachmentAPIResponse{}

		err = json.Unmarshal(bodyData, &res2)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", res2.Url, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		if !reflect.DeepEqual(newAttachment, bodyResponseBytes) {
			t.Errorf("Attachment content mismatch")
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Attachments) != 1 {
			t.Error(ErrorMismatch("len(Attachments)", fmt.Sprint(len(meta.Attachments)), fmt.Sprint(1)))
		}

		if meta.Attachments[0].Name != "test-attachment.txt" {
			t.Error(ErrorMismatch("Attachment name", meta.Attachments[0].Name, "test-attachment.txt"))
		}

		if meta.Attachments[0].Size != uint64(len(newAttachment)) {
			t.Error(ErrorMismatch("Attachment size", fmt.Sprint(meta.Attachments[0].Size), fmt.Sprint(len(newAttachment))))
		}

		// Rename attachment

		reqBody, err := json.Marshal(MediaAttachmentEditNameBody{
			Id:   meta.Attachments[0].Id,
			Name: "new-name.log",
		})

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/attachments/rename", reqBody, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Attachments) != 1 {
			t.Error(ErrorMismatch("len(Attachments)", fmt.Sprint(len(meta.Attachments)), fmt.Sprint(1)))
		}

		if meta.Attachments[0].Name != "new-name.log" {
			t.Error(ErrorMismatch("Attachment name", meta.Attachments[0].Name, "new-name.log"))
		}

		if meta.Attachments[0].Size != uint64(len(newAttachment)) {
			t.Error(ErrorMismatch("Attachment size", fmt.Sprint(meta.Attachments[0].Size), fmt.Sprint(len(newAttachment))))
		}

		// Delete attachment

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/attachments/remove?id="+fmt.Sprint(meta.Attachments[0].Id), nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Attachments) != 0 {
			t.Error(ErrorMismatch("len(Attachments)", fmt.Sprint(len(meta.Attachments)), fmt.Sprint(0)))
		}
	}

	// Extra resolutions

	if meta.Type == MediaTypeImage || meta.Type == MediaTypeVideo {
		testResolution := ApiMediaResolutionBody{
			Width:  256,
			Height: 144,
			Fps:    30,
		}

		if meta.Type == MediaTypeImage {
			testResolution.Fps = 1
		}

		body, err = json.Marshal(testResolution)

		if err != nil {
			t.Error(err)
			return
		}

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/resolution/add", body, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Resolutions) != 1 {
			t.Error(ErrorMismatch("len(Resolutions)", fmt.Sprint(len(meta.Resolutions)), "1"))
			return
		}

		if meta.Resolutions[0].Width != testResolution.Width || meta.Resolutions[0].Height != testResolution.Height || meta.Resolutions[0].Fps != testResolution.Fps {
			t.Errorf("Added resolution does not match the existing one")
		}

		for !meta.Resolutions[0].Ready {
			if meta.Resolutions[0].Task > 0 {
				statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/tasks/"+fmt.Sprint(meta.Resolutions[0].Task), nil, session)

				if err != nil {
					t.Error(err)
					return
				}

				if statusCode == 200 {
					taskInfo := TaskListInfoEntry{}

					err = json.Unmarshal(bodyResponseBytes, &taskInfo)

					if err != nil {
						t.Error(err)
						return
					}
				} else if statusCode != 404 {
					t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200/404"))
				}
			}

			time.Sleep(100 * time.Millisecond)

			err = _TestFetchMetadata(server, session, t, mediaId, &meta)

			if err != nil {
				t.Error(err)
				return
			}

			if len(meta.Resolutions) != 1 {
				t.Error(ErrorMismatch("len(Resolutions)", fmt.Sprint(len(meta.Resolutions)), "1"))
				return
			}
		}

		statusCode, _, err = DoTestRequest(server, "GET", meta.Resolutions[0].Url, nil, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/resolution/remove", body, session)

		if err != nil {
			t.Error(err)
			return
		}

		if statusCode != 200 {
			t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
		}

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}

		if len(meta.Resolutions) != 0 {
			t.Error(ErrorMismatch("len(Resolutions)", fmt.Sprint(len(meta.Resolutions)), "0"))
			return
		}
	}

	// Re-Encode

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/encode", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	err = _TestFetchMetadata(server, session, t, mediaId, &meta)

	if err != nil {
		t.Error(err)
		return
	}

	for !meta.Encoded {
		if meta.Task > 0 {
			statusCode, bodyResponseBytes, err = DoTestRequest(server, "GET", "/api/tasks/"+fmt.Sprint(meta.Task), nil, session)

			if err != nil {
				t.Error(err)
				return
			}

			if statusCode == 200 {
				taskInfo := TaskListInfoEntry{}

				err = json.Unmarshal(bodyResponseBytes, &taskInfo)

				if err != nil {
					t.Error(err)
					return
				}
			} else if statusCode != 404 {
				t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200/404"))
			}
		}

		time.Sleep(100 * time.Millisecond)

		err = _TestFetchMetadata(server, session, t, mediaId, &meta)

		if err != nil {
			t.Error(err)
			return
		}
	}

	// HEAD request to asset

	statusCode, originalAssetBytes, err := DoTestRequest(server, "GET", meta.Url, nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, resHead, _, err := DoTestRangeRequest(server, session, "HEAD", meta.Url, "")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	assetBytes, err := strconv.ParseInt(resHead.Get("Content-Length"), 10, 64)

	if err != nil {
		t.Error(err)
		return
	}

	if assetBytes != int64(len(originalAssetBytes)) {
		t.Error(ErrorMismatch("HeadContentLength", fmt.Sprint(assetBytes), fmt.Sprint(len(originalAssetBytes))))
	}

	// Range request (0-X)

	statusCode, resHead, bodyResponseBytes, err = DoTestRangeRequest(server, session, "GET", meta.Url, "bytes=0-31")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 206 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "206"))
	}

	if resHead.Get("Content-Length") != "32" {
		t.Error(ErrorMismatch("Content-Length", resHead.Get("Content-Length"), "32"))
	}

	expectedContentRange := "bytes " + fmt.Sprint(0) + "-" + fmt.Sprint(31) + "/" + fmt.Sprint(assetBytes)

	if resHead.Get("Content-Range") != expectedContentRange {
		t.Error(ErrorMismatch("Content-Range", resHead.Get("Content-Range"), expectedContentRange))
	}

	if !reflect.DeepEqual(bodyResponseBytes, originalAssetBytes[0:32]) {
		t.Errorf("Asset bytes mismatch")
	}

	// Range request (X-Y)

	statusCode, resHead, bodyResponseBytes, err = DoTestRangeRequest(server, session, "GET", meta.Url, "bytes=32-63")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 206 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "206"))
	}

	if resHead.Get("Content-Length") != "32" {
		t.Error(ErrorMismatch("Content-Length", resHead.Get("Content-Length"), "32"))
	}

	expectedContentRange = "bytes " + fmt.Sprint(32) + "-" + fmt.Sprint(63) + "/" + fmt.Sprint(assetBytes)

	if resHead.Get("Content-Range") != expectedContentRange {
		t.Error(ErrorMismatch("Content-Range", resHead.Get("Content-Range"), expectedContentRange))
	}

	if !reflect.DeepEqual(bodyResponseBytes, originalAssetBytes[32:64]) {
		t.Errorf("Asset bytes mismatch")
	}

	// Range request (X-)

	statusCode, resHead, bodyResponseBytes, err = DoTestRangeRequest(server, session, "GET", meta.Url, "bytes=32-")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 206 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "206"))
	}

	if resHead.Get("Content-Length") != fmt.Sprint(assetBytes-32) {
		t.Error(ErrorMismatch("Content-Length", resHead.Get("Content-Length"), fmt.Sprint(assetBytes-32)))
	}

	expectedContentRange = "bytes " + fmt.Sprint(32) + "-" + fmt.Sprint(assetBytes-1) + "/" + fmt.Sprint(assetBytes)

	if resHead.Get("Content-Range") != expectedContentRange {
		t.Error(ErrorMismatch("Content-Range", resHead.Get("Content-Range"), expectedContentRange))
	}

	if !reflect.DeepEqual(bodyResponseBytes, originalAssetBytes[32:]) {
		t.Errorf("Asset bytes mismatch")
	}

	// Range request (-Y)

	statusCode, resHead, bodyResponseBytes, err = DoTestRangeRequest(server, session, "GET", meta.Url, "bytes=-32")

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 206 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "206"))
	}

	if resHead.Get("Content-Length") != fmt.Sprint(32) {
		t.Error(ErrorMismatch("Content-Length", resHead.Get("Content-Length"), "32"))
	}

	expectedContentRange = "bytes " + fmt.Sprint(assetBytes-32) + "-" + fmt.Sprint(assetBytes-1) + "/" + fmt.Sprint(assetBytes)

	if resHead.Get("Content-Range") != expectedContentRange {
		t.Error(ErrorMismatch("Content-Range", resHead.Get("Content-Range"), expectedContentRange))
	}

	if !reflect.DeepEqual(bodyResponseBytes, originalAssetBytes[assetBytes-32:]) {
		t.Error(ErrorMismatch("Content-Length", fmt.Sprint(bodyResponseBytes), fmt.Sprint(originalAssetBytes[assetBytes-32:])))
	}

	// Finish: Delete media

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId))+"/delete", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}
}

func UploadTestMedia(server *httptest.Server, session string, mt MediaType, title string, album string) (mid uint64, e error) {
	client := server.Client()

	fileName := ""

	switch mt {
	case MediaTypeVideo:
		fileName = "test-video.mp4"
	case MediaTypeAudio:
		fileName = "test-audio.mp3"
	case MediaTypeImage:
		fileName = "test-image.png"
	default:
		return 0, errors.New("Invalid media type")
	}

	apiURL, err := url.JoinPath(server.URL, "/api/upload")

	if err != nil {
		return 0, err
	}

	var b bytes.Buffer

	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile("file", fileName)

	if err != nil {
		return 0, err
	}

	fileData, err := os.ReadFile(path.Join("test-assets", fileName))

	if err != nil {
		return 0, err
	}

	_, err = part.Write(fileData)

	if err != nil {
		return 0, err
	}

	writer.Close()

	req, err := http.NewRequest("POST", apiURL+"?title="+url.QueryEscape(title)+"&album="+url.QueryEscape(album), &b)

	if err != nil {
		return 0, err
	}

	req.Header.Set("x-session-token", session)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("Upload error. Status code: " + fmt.Sprint(resp.StatusCode))
	}

	bodyData, err := io.ReadAll(resp.Body)

	if err != nil {
		return 0, err
	}

	uploadRes := UploadAPIResponse{}

	err = json.Unmarshal(bodyData, &uploadRes)

	if err != nil {
		return 0, err
	}

	mediaId := uploadRes.Id

	// Wait for the media to be ready

	mediaReady := false

	for !mediaReady {
		statusCode, bodyResponseBytes, err := DoTestRequest(server, "GET", "/api/media/"+url.PathEscape(fmt.Sprint(mediaId)), nil, session)

		if err != nil {
			return 0, err
		}

		if statusCode != 200 {
			return 0, ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200")
		}

		meta := MediaAPIMetaResponse{}

		err = json.Unmarshal(bodyResponseBytes, &meta)

		if err != nil {
			return 0, err
		}

		if meta.Type != mt {
			return 0, ErrorMismatch("MediaType", fmt.Sprint(meta.Type), fmt.Sprint(mt))
		}

		mediaReady = meta.Ready

		if !mediaReady {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return mediaId, nil
}
