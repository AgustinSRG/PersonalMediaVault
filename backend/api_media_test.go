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
	"sync"
	"testing"
	"time"
)

func Media_API_Test(server *httptest.Server, session string, t *testing.T) {
	// Upload media and wait for it to be ready
	wg := sync.WaitGroup{}

	var videoId uint64
	var audioId uint64
	var imageId uint64
	var isError bool = false

	wg.Add(1)
	t.Run("Upload video", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeVideo, "Test video", "")

		if err != nil {
			t.Error(err)
			isError = true
		}

		videoId = mid

		wg.Done()
	})

	wg.Add(1)
	t.Run("Upload audio", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeAudio, "Test audio", "")

		if err != nil {
			t.Error(err)
			isError = true
		}

		audioId = mid

		wg.Done()
	})

	wg.Add(1)
	t.Run("Upload image", func(t *testing.T) {
		mid, err := UploadTestMedia(server, session, MediaTypeImage, "Test image", "")

		if err != nil {
			t.Error(err)
			isError = true
		}

		imageId = mid

		wg.Done()
	})

	wg.Wait()

	if isError {
		return
	}

	// Delete media

	statusCode, _, err := DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(videoId))+"/delete", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(audioId))+"/delete", nil, session)

	if err != nil {
		t.Error(err)
		return
	}

	if statusCode != 200 {
		t.Error(ErrorMismatch("StatusCode", fmt.Sprint(statusCode), "200"))
	}

	statusCode, _, err = DoTestRequest(server, "POST", "/api/media/"+url.PathEscape(fmt.Sprint(imageId))+"/delete", nil, session)

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
