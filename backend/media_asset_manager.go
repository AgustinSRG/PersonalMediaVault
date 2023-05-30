// Media assset manager

package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
)

// Data stored in disk needed to manage the media assets
type MediaAssetManagerData struct {
	NextId uint64 `json:"next_id"` // ID to assign to the next media asset
}

// Manages the media assets of the vault
type MediaAssetsManager struct {
	path string // Vault base path

	data_file      string      // Path to the manager data file (stored MediaAssetManagerData)
	data_file_lock *sync.Mutex // Mutex to access the data file

	assets map[uint64]*MediaAsset // Mapping of active media assets (assets that are being used)

	lock *sync.Mutex // Mutex to control access to the assets mapping

	ready_progress_map  map[uint64]int32 // Mapping to store upload progress
	ready_progress_lock *sync.Mutex      // Mutex to control acess to ready_progress_map
}

// Manager initialization
// base_path - Vault base path
func (mm *MediaAssetsManager) Initialize(base_path string) {
	mm.lock = &sync.Mutex{}

	mm.path = base_path

	mm.data_file = path.Join(base_path, "media_ids.json")
	mm.data_file_lock = &sync.Mutex{}

	mm.ready_progress_map = make(map[uint64]int32)
	mm.ready_progress_lock = &sync.Mutex{}

	mm.assets = make(map[uint64]*MediaAsset)
}

// Read manager data
// Note: This is a thread-unsafe internal method
func (mm *MediaAssetsManager) readData() (*MediaAssetManagerData, error) {
	if _, err := os.Stat(mm.data_file); err == nil {
		// Load file
		bJSON, err := os.ReadFile(mm.data_file)

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

	err = os.WriteFile(tmpFile, jsonData, FILE_PERMISSION)

	if err != nil {
		return 0, err
	}

	// Save to original file

	err = RenameAndReplace(tmpFile, mm.data_file)

	if err != nil {
		return 0, err
	}

	return mediaId, nil
}

// Resolves the path where a media asset is stored
// media_id - Media ID
func (mm *MediaAssetsManager) ResolveMediaPath(media_id uint64) string {
	prefixByte := byte(media_id % 256)
	prefixByteHex := hex.EncodeToString([]byte{prefixByte})

	return path.Join(mm.path, "media", prefixByteHex, fmt.Sprint(media_id))
}

// Acquires a media resource, creating an struct to manage its status
// media_id - Media ID
// Returns a reference to MediaAsset
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

// Releases resources created by AcquireMediaResource()
// media_id - Media ID
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

// Gets the upload progress for a media asset
// mid - Media ID
// Returns the progress (0 - 100)
func (mm *MediaAssetsManager) GetProgress(mid uint64) int32 {
	mm.ready_progress_lock.Lock()
	defer mm.ready_progress_lock.Unlock()

	return mm.ready_progress_map[mid]
}

// Sets the upload progress for a media asset
// mid - Media ID
// p - Progress (0 - 100)
func (mm *MediaAssetsManager) SetProgress(mid uint64, p int32) {
	mm.ready_progress_lock.Lock()
	defer mm.ready_progress_lock.Unlock()

	mm.ready_progress_map[mid] = p
}

// Clears the upload progress for a media asset
// Call after it's fully uploaded
// mid - Media ID
func (mm *MediaAssetsManager) EndProgress(mid uint64) {
	mm.ready_progress_lock.Lock()
	defer mm.ready_progress_lock.Unlock()

	delete(mm.ready_progress_map, mid)
}
