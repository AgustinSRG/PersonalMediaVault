// Media assset manager

package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

type MediaAssetManagerData struct {
	NextId uint64 `json:"next_id"`
}

type MediaAssetsManager struct {
	path string

	data_file      string
	data_file_lock *sync.Mutex

	assets map[uint64]*MediaAsset

	lock *sync.Mutex
}

func (mm *MediaAssetsManager) Initialize(base_path string) {
	mm.lock = &sync.Mutex{}

	mm.path = base_path

	mm.data_file = path.Join(base_path, "media_ids.json")
	mm.data_file_lock = &sync.Mutex{}

	mm.assets = make(map[uint64]*MediaAsset)
}

func (mm *MediaAssetsManager) readData() (*MediaAssetManagerData, error) {
	if _, err := os.Stat(mm.data_file); err == nil {
		// Load file
		bJSON, err := ioutil.ReadFile(mm.data_file)

		if err != nil {
			return nil, err
		}

		var mp MediaAssetManagerData

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No albums yet

		mp := MediaAssetManagerData{
			NextId: 0,
		}

		return &mp, nil
	} else {
		return nil, err
	}
}

// Generates a new ID for a new media asset
// Makes sure it's saved, so it's unique between restarts
func (mm *MediaAssetsManager) NextMediaId() (uint64, error) {
	mm.data_file_lock.Lock()
	defer mm.data_file_lock.Unlock()

	data, err := mm.readData()

	if err != nil {
		return 0, err
	}

	mediaId := data.NextId

	data.NextId++

	// Save

	jsonData, err := json.Marshal(data)

	if err != nil {
		return 0, err
	}

	// Create temp file to write it

	tmpFile := GetTemporalFileName("json", true)

	err = ioutil.WriteFile(tmpFile, jsonData, FILE_PERMISSION)

	if err != nil {
		return 0, err
	}

	// Save to original file

	err = os.Rename(tmpFile, mm.data_file)

	if err != nil {
		return 0, err
	}

	return mediaId, nil
}

func (mm *MediaAssetsManager) ResolveMediaPath(media_id uint64) string {
	prefixByte := byte(media_id % 256)
	prefixByteHex := hex.EncodeToString([]byte{prefixByte})

	return path.Join(mm.path, "media", prefixByteHex, fmt.Sprint(media_id))
}

func (mm *MediaAssetsManager) AcquireMediaResource(media_id uint64) *MediaAsset {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	if mm.assets[media_id] != nil {
		mm.assets[media_id].use_count++
		return mm.assets[media_id]
	}

	newAsset := MediaAsset{
		id:        media_id,
		path:      mm.ResolveMediaPath(media_id),
		lock:      CreateReadWriteLock(),
		use_count: 1,
		mu:        &sync.Mutex{},
		files:     make(map[uint64]*MediaAssetFile),
		deleting:  false,
	}

	mm.assets[media_id] = &newAsset

	return &newAsset
}

func (mm *MediaAssetsManager) ReleaseMediaResource(media_id uint64) {
	mm.lock.Lock()
	defer mm.lock.Unlock()

	if mm.assets[media_id] == nil {
		return
	}

	mm.assets[media_id].use_count--

	if mm.assets[media_id].use_count <= 0 {
		delete(mm.assets, media_id)
	}
}
