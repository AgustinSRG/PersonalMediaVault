// Media asset

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
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

	deleting bool

	mu *sync.Mutex

	files map[uint64]*MediaAssetFile
}

type MediaAssetFile struct {
	id uint64

	lock *ReadWriteLock

	use_count int32

	waiting bool
	wait_mu *sync.Mutex
}

type MediaResolution struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	Fps    int32 `json:"fps"`

	Ready     bool   `json:"ready"`
	Asset     uint64 `json:"asset"`
	Extension string `json:"ext"`

	TaskId uint64 `json:"task_id"`
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

	OriginalReady     bool   `json:"original_ready"`
	OriginalAsset     uint64 `json:"original_asset"`
	OriginalExtension string `json:"original_ext"`
	OriginalTask      uint64 `json:"original_task"`
	OriginalEncoded   bool   `json:"original_encoded"`

	ThumbnailReady bool   `json:"thumb_ready"`
	ThumbnailAsset uint64 `json:"thumb_asset"`

	Resolutions []MediaResolution `json:"resolutions"`

	PreviewsReady    bool    `json:"previews_ready"`
	PreviewsTask     uint64  `json:"previews_task"`
	PreviewsInterval float64 `json:"previews_interval"`
	PreviewsAsset    uint64  `json:"previews_asset"`
}

func (media *MediaAsset) CreateNewMediaAsset(key []byte, media_type MediaType, title string, desc string, duration float64, width int32, height int32) error {
	now := time.Now().UnixMilli()

	meta := MediaMetadata{
		Id:                media.id,
		Type:              media_type,
		MediaDuration:     duration,
		Width:             width,
		Height:            height,
		Title:             title,
		Description:       desc,
		Tags:              make([]uint64, 0),
		UploadTimestamp:   now,
		NextAssetID:       0,
		OriginalReady:     false,
		OriginalAsset:     0,
		OriginalTask:      0,
		OriginalEncoded:   false,
		OriginalExtension: "",
		ThumbnailReady:    false,
		ThumbnailAsset:    0,
		Resolutions:       make([]MediaResolution, 0),
		PreviewsReady:     false,
		PreviewsInterval:  0,
		PreviewsAsset:     0,
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

const (
	ASSET_MUTI_FILE   = "m"
	ASSET_SINGLE_FILE = "s"
)

func (media *MediaAsset) GetAssetPath(asset_id uint64, asset_type string) string {
	return path.Join(media.path, asset_type+"_"+fmt.Sprint(asset_id)+".pma")
}

func (media *MediaAsset) AcquireAsset(asset_id uint64, asset_type string) (bool, string, *ReadWriteLock) {
	media.mu.Lock()
	defer media.mu.Unlock()

	if media.deleting {
		return false, "", nil
	}

	p := media.GetAssetPath(asset_id, asset_type)

	if media.files[asset_id] != nil {
		media.files[asset_id].use_count++
		return true, p, media.files[asset_id].lock
	}

	f := MediaAssetFile{
		id:        asset_id,
		lock:      CreateReadWriteLock(),
		use_count: 1,
		waiting:   false,
		wait_mu:   &sync.Mutex{},
	}

	media.files[asset_id] = &f

	return true, p, f.lock
}

func (media *MediaAsset) ReleaseAsset(asset_id uint64) {
	media.mu.Lock()
	defer media.mu.Unlock()

	if media.files[asset_id] != nil {
		media.files[asset_id].use_count--

		if media.files[asset_id].use_count <= 0 {

			if media.files[asset_id].waiting {
				media.files[asset_id].wait_mu.Unlock()
			}

			delete(media.files, asset_id)
		}
	}
}

func (media *MediaAsset) Delete() {
	// Delete metadata file

	media.lock.RequestWrite()
	media.lock.StartWrite()

	os.Remove(path.Join(media.path, "meta.pmv"))

	media.lock.EndWrite()

	go media.deleteAll()
}

func (media *MediaAsset) deleteAll() {
	// Set deleting and wait for assets to be released

	locks := make([]*sync.Mutex, 0)

	media.mu.Lock()

	if media.deleting {
		return
	}

	media.deleting = true

	for _, a := range media.files {
		a.waiting = true
		a.wait_mu.Lock()
		locks = append(locks, a.wait_mu)
	}

	media.mu.Unlock()

	for i := 0; i < len(locks); i++ {
		locks[i].Lock()
		locks[i].Unlock()
	}

	// Now, delete evrything

	media.lock.RequestWrite()
	media.lock.StartWrite()

	os.RemoveAll(media.path)

	media.lock.EndWrite()
}
