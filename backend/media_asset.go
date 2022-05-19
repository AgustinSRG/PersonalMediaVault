// Media asset

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type MediaType uint16

const (
	MediaTypeImage MediaType = 1
	MediaTypeVideo MediaType = 2
	MediaTypeAudio MediaType = 3
)

type MediaAsset struct {
	id uint64

	path string
	lock *ReadWriteLock

	use_count int32
}

type MediaResolution struct {
	Original bool `json:"original"`

	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	Fps    int32 `json:"fps"`

	Ready bool   `json:"ready"`
	Asset uint64 `json:"asset"`

	TaskCreated bool   `json:"task_created"`
	TaskId      uint64 `json:"task_id"`
	TaskError   string `json:"task_error"`
}

type MediaMetadata struct {
	Id uint64 `json:"id"`

	Type MediaType `json:"type"`

	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []uint64 `json:"tags"`

	MediaDuration float64 `json:"duration"`
	Width         int32   `json:"width"`
	Height        int32   `json:"height"`

	UploadTimestamp int64 `json:"upload_time"`

	NextAssetID uint64 `json:"next_asset_id"`

	OriginalReady bool   `json:"original_ready"`
	OriginalAsset uint64 `json:"original_asset"`

	ThumbnailReady bool   `json:"thumb_ready"`
	ThumbnailAsset uint64 `json:"thumb_asset"`

	Resolutions []MediaResolution `json:"resolutions"`

	PreviewsReady    bool    `json:"previews_ready"`
	PreviewsInterval float64 `json:"previews_interval"`
	PreviewsError    string  `json:"previews_error"`
	PreviewsAsset    uint64  `json:"previews_asset"`
}

func (media *MediaAsset) CreateNewMediaAsset(key []byte, media_type MediaType, title string, desc string, duration float64, width int32, height int32) error {
	now := time.Now().UnixMilli()

	meta := MediaMetadata{
		Id:               media.id,
		Type:             media_type,
		MediaDuration:    duration,
		Width:            width,
		Height:           height,
		Title:            title,
		Description:      desc,
		Tags:             make([]uint64, 0),
		UploadTimestamp:  now,
		NextAssetID:      0,
		OriginalReady:    false,
		OriginalAsset:    0,
		ThumbnailReady:   false,
		ThumbnailAsset:   0,
		Resolutions:      make([]MediaResolution, 0),
		PreviewsReady:    false,
		PreviewsInterval: 0,
		PreviewsError:    "",
		PreviewsAsset:    0,
	}

	media.lock.RequestWrite() // Request write
	defer media.lock.EndWrite()

	jsonData, err := json.Marshal(meta)

	if err != nil {
		return err
	}

	encData, err := encryptFileContents(jsonData, AES256_ZIP, key)

	if err != nil {
		return err
	}

	// Create temp file to write it

	tmpFile := GetTemporalFileName("pmv")

	err = ioutil.WriteFile(tmpFile, encData, FILE_PERMISSION)

	if err != nil {
		return err
	}

	// Save to original file
	media.lock.StartWrite()

	err = os.Rename(tmpFile, path.Join(media.path, "meta.pmv"))

	return err
}

func (media *MediaAsset) readData(key []byte) (*MediaMetadata, error) {
	file := path.Join(media.path, "meta.pmv")
	if _, err := os.Stat(file); err == nil {
		// Load file
		b, err := ioutil.ReadFile(file)

		if err != nil {
			return nil, err
		}

		bJSON, err := decryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp MediaMetadata

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No tags

		return nil, nil
	} else {
		return nil, err
	}
}

func (media *MediaAsset) ReadMeatadata(key []byte) (*MediaMetadata, error) {
	media.lock.StartRead() // Request read
	defer media.lock.EndRead()

	return media.readData(key)
}

func (media *MediaAsset) StartWrite(key []byte) (*MediaMetadata, error) {
	media.lock.RequestWrite() // Request write

	return media.readData(key)
}

func (media *MediaAsset) EndWrite(data *MediaMetadata, key []byte) error {
	defer media.lock.EndWrite()

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	encData, err := encryptFileContents(jsonData, AES256_ZIP, key)

	if err != nil {
		return err
	}

	// Create temp file to write it

	tmpFile := GetTemporalFileName("pmv")

	err = ioutil.WriteFile(tmpFile, encData, FILE_PERMISSION)

	if err != nil {
		return err
	}

	// Save to original file
	media.lock.StartWrite()

	err = os.Rename(tmpFile, path.Join(media.path, "meta.pmv"))

	return err
}

func (media *MediaAsset) CancelWrite() {
	media.lock.EndWrite()
}
