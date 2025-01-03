// Album manager

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

const ALBUM_MAX_SIZE = 1024

var ErrAlbumMaxSizeReached = errors.New("max size reached for the album")

// Album data
type VaultAlbumData struct {
	Name         string   `json:"name"`  // Name of the album
	List         []uint64 `json:"list"`  // Ordered list of media to play
	LastModified int64    `json:"lm"`    // Last modified timestamp
	Thumbnail    *uint64  `json:"thumb"` // Thumbnail asset
}

// Album list data
type VaultAlbumsData struct {
	NextId               uint64                     `json:"next_id"`                 // Id for the next album to create
	NextThumbnailAssetId uint64                     `json:"next_thumb_id,omitempty"` // Id for the next album thumbnail asset
	Albums               map[uint64]*VaultAlbumData `json:"albums"`                  // Albums (Id -> Data)
}

// Checks if an album has a media in it
// media_id - Id of the media file
// Returns true if the album has the media in its list
func (data *VaultAlbumData) HasMedia(media_id uint64) bool {
	for i := 0; i < len(data.List); i++ {
		if data.List[i] == media_id {
			return true
		}
	}

	return false
}

// Represents an asset file that stores a thumbnail of an album
type AlbumThumbnailAsset struct {
	id uint64 // File ID

	lock *ReadWriteLock // Lock to control read/write operations

	use_count int32 // Counter of threads accessing the file
}

// Album manager
type VaultAlbumsManager struct {
	path        string // Vault path
	albums_file string // Path to the albums data file

	cache *VaultAlbumsData // Cache
	lock  *ReadWriteLock   // Lock to control access to the file

	thumbnails   map[uint64]*AlbumThumbnailAsset // Files that the media asset has
	thumbnailsMu *sync.Mutex

	thumbnail_cache *ThumbnailCache // Thumbnail cache
}

// Initializes albums manager
// base_path - Vault path
func (am *VaultAlbumsManager) Initialize(base_path string) {
	am.path = base_path
	am.albums_file = path.Join(base_path, "albums.pmv")

	am.cache = nil

	am.lock = CreateReadWriteLock()

	am.thumbnails = make(map[uint64]*AlbumThumbnailAsset)
	am.thumbnailsMu = &sync.Mutex{}

	am.thumbnail_cache = makeThumbnailCache()

	// Make thumbnails folder if not present
	_ = os.MkdirAll(path.Join(base_path, "thumb_album"), FOLDER_PERMISSION)
}

// Reads albums list data
// key - Vault decryption key
// Returns the data
// Thread unsafe: This is an internal method
func (am *VaultAlbumsManager) readData(key []byte) (*VaultAlbumsData, error) {
	if am.cache != nil {
		return am.cache, nil
	}

	if _, err := os.Stat(am.albums_file); err == nil {
		// Load file
		b, err := os.ReadFile(am.albums_file)

		if err != nil {
			return nil, err
		}

		bJSON, err := encrypted_storage.DecryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp VaultAlbumsData

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		if mp.Albums == nil {
			mp.Albums = make(map[uint64]*VaultAlbumData)
		}

		am.cache = &mp

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No albums yet

		mp := VaultAlbumsData{
			NextId:               0,
			NextThumbnailAssetId: 0,
			Albums:               make(map[uint64]*VaultAlbumData),
		}

		am.cache = &mp

		return &mp, nil
	} else {
		return nil, err
	}
}

// Reads albums data
// key - Vault decryption key
// Returns the albums data
func (am *VaultAlbumsManager) ReadAlbums(key []byte) (*VaultAlbumsData, error) {
	am.lock.StartRead() // Request read
	defer am.lock.EndRead()

	return am.readData(key)
}

// Starts a write operation in the albums data
// key - Vault decryption key
// Returns the albums data
// After calling this, the file is locked until you call EndWrite() or CancelWrite()
func (am *VaultAlbumsManager) StartWrite(key []byte) (*VaultAlbumsData, error) {
	am.lock.RequestWrite() // Request write

	var data_copy VaultAlbumsData

	data, err := am.readData(key)

	if err != nil {
		return nil, err
	}

	data_copy = *data

	return &data_copy, nil
}

// Applies changes to albums data
// data - Data to write
// key - Vault encryption key
func (am *VaultAlbumsManager) EndWrite(data *VaultAlbumsData, key []byte) error {
	defer am.lock.EndWrite()

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	encData, err := encrypted_storage.EncryptFileContents(jsonData, encrypted_storage.AES256_ZIP, key)

	if err != nil {
		return err
	}

	// Create temp file to write it

	tmpFile := GetTemporalFileName("pmv", true)

	err = os.WriteFile(tmpFile, encData, FILE_PERMISSION)

	if err != nil {
		return err
	}

	// Save to original file
	am.lock.StartWrite()

	err = RenameAndReplace(tmpFile, am.albums_file)

	am.cache = data

	return err
}

// Cancels pending write operation
func (am *VaultAlbumsManager) CancelWrite() {
	am.lock.EndWrite()
}

// Creates new album
// name - Name for the album
// key - Vault encryption key
// Returns the ID for the new album
func (am *VaultAlbumsManager) CreateAlbum(name string, key []byte) (uint64, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return 0, err
	}

	album_id := data.NextId

	data.NextId++

	data.Albums[album_id] = &VaultAlbumData{
		Name:         name,
		List:         make([]uint64, 0),
		LastModified: time.Now().UnixMilli(),
	}

	err = am.EndWrite(data, key)

	return album_id, err
}

// Gets ID for a thumbnail asset
// key - Vault encryption key
// Returns the ID of the asset if success
func (am *VaultAlbumsManager) GetThumbnailAssetId(key []byte) (uint64, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return 0, err
	}

	id := data.NextThumbnailAssetId

	data.NextThumbnailAssetId++

	err = am.EndWrite(data, key)

	return id, err
}

// Adds a media file to an album
// album_id - Album ID
// media_id - Media file ID
// key - Vault encryption key
// Returns true if the media was added, false if the media was already in the list
func (am *VaultAlbumsManager) AddMediaToAlbum(album_id uint64, media_id uint64, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	if data.Albums[album_id].HasMedia(media_id) {
		am.CancelWrite()
		return true, nil // Already added
	}

	old_list := data.Albums[album_id].List

	if len(old_list) >= ALBUM_MAX_SIZE {
		am.CancelWrite()
		return false, ErrAlbumMaxSizeReached
	}

	new_list := append(old_list, media_id)

	data.Albums[album_id].List = new_list
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	if len(old_list) == 0 {
		am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)
	}

	return true, err
}

// Removes a media file from an album
// album_id - Album ID
// media_id - Media file ID
// key - Vault encryption key
// Returns true if the media was removed, false if the media was not in the list
func (am *VaultAlbumsManager) RemoveMediaFromAlbum(album_id uint64, media_id uint64, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	if !data.Albums[album_id].HasMedia(media_id) {
		am.CancelWrite()
		return true, nil // Not found
	}

	old_list := data.Albums[album_id].List

	first_changed := len(data.Albums[album_id].List) > 0 && data.Albums[album_id].List[0] == media_id

	new_list := make([]uint64, 0)

	for i := 0; i < len(old_list); i++ {
		if old_list[i] != media_id {
			new_list = append(new_list, old_list[i])
		}
	}

	data.Albums[album_id].List = new_list
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	if first_changed {
		am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)
	}

	return true, err
}

// Removes a media file from an album
// album_id - Album ID
// media_id - Media file ID
// key - Vault encryption key
// Returns true if the media was moved, false if the media was not moved
func (am *VaultAlbumsManager) MoveMediaToPositionInAlbum(album_id uint64, media_id uint64, position int, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	old_list := data.Albums[album_id].List

	if !data.Albums[album_id].HasMedia(media_id) {
		if len(old_list) >= ALBUM_MAX_SIZE {
			am.CancelWrite()
			return false, ErrAlbumMaxSizeReached
		}
	}

	if position < 0 {
		position = 0
	}

	if position > len(old_list) {
		position = len(old_list)
	}

	first_changed := position == 0

	new_list := make([]uint64, 0)

	j := 0 // Position in the new list

	for i := 0; i < len(old_list); i++ {
		if j == position {
			new_list = append(new_list, media_id)
			j++
		}
		if old_list[i] != media_id {
			new_list = append(new_list, old_list[i])
			j++
		}
	}

	if position >= len(new_list) {
		new_list = append(new_list, media_id)
	}

	data.Albums[album_id].List = new_list
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	if first_changed {
		am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)
	}

	return true, err
}

// Sets album media list
// album_id - Album ID
// media_list - List of media files
// key - Vault encryption key
// Returns true if success
func (am *VaultAlbumsManager) SetAlbumList(album_id uint64, media_list []uint64, key []byte) (bool, error) {
	if len(media_list) > ALBUM_MAX_SIZE {
		return false, ErrAlbumMaxSizeReached
	}

	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	data.Albums[album_id].List = AlbumListPruneRepeatedElements(media_list)
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)

	return true, err
}

// Sets album name
// album_id - Album ID
// name - Album name
// key - Vault encryption key
// Returns true if success
func (am *VaultAlbumsManager) RenameAlbum(album_id uint64, name string, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	data.Albums[album_id].Name = name
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	return true, err
}

// Sets album name
// album_id - Album ID
// thumb_asset_id - Thumbnail asset ID
// key - Vault encryption key
// Returns true if success, and possibly the ID of the old thumbnail asset
func (am *VaultAlbumsManager) SetAlbumThumbnail(album_id uint64, thumb_asset_id uint64, key []byte) (bool, *uint64, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, nil, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil, nil // Not found
	}

	old_asset := data.Albums[album_id].Thumbnail

	data.Albums[album_id].Thumbnail = &thumb_asset_id
	data.Albums[album_id].LastModified = time.Now().UnixMilli()

	err = am.EndWrite(data, key)

	am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)

	return true, old_asset, err
}

// Deletes album
// album_id - Album ID
// key - Vault encryption key
func (am *VaultAlbumsManager) DeleteAlbum(album_id uint64, key []byte) error {
	data, err := am.StartWrite(key)

	if err != nil {
		return err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return nil // Not found
	}

	delete(data.Albums, album_id)

	err = am.EndWrite(data, key)

	am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)

	return err
}

// Called on media delete to remove it from all albums
// media_id - Media file ID
// key - Vault encryption key
func (am *VaultAlbumsManager) OnMediaDelete(media_id uint64, key []byte) error {
	data, err := am.StartWrite(key)

	if err != nil {
		return err
	}

	album_thumbnail_caches_to_remove := make([]uint64, 0)

	for album_id := range data.Albums {
		if !data.Albums[album_id].HasMedia(media_id) {
			continue
		}

		old_list := data.Albums[album_id].List
		new_list := make([]uint64, 0)

		for i := 0; i < len(old_list); i++ {
			if old_list[i] != media_id {
				new_list = append(new_list, old_list[i])
			}
		}

		data.Albums[album_id].List = new_list

		album_thumbnail_caches_to_remove = append(album_thumbnail_caches_to_remove, album_id)
	}

	err = am.EndWrite(data, key)

	for i := 0; i < len(album_thumbnail_caches_to_remove); i++ {
		am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_thumbnail_caches_to_remove[i])
	}

	return err
}

// Called on media thumbnail update to update
// media_id - Media file ID
// key - Vault encryption key
func (am *VaultAlbumsManager) OnMediaThumbnailUpdate(media_id uint64, key []byte) error {
	data, err := am.ReadAlbums(key)

	if err != nil {
		return err
	}

	for album_id := range data.Albums {
		if len(data.Albums[album_id].List) == 0 {
			continue
		}

		if data.Albums[album_id].List[0] != media_id {
			continue
		}

		am.thumbnail_cache.RemoveEntryOrMarkInvalid(album_id)
	}

	return nil
}

// Reads the albums and pre-caches them on vault unlock
// key - Vault decryption key
func (am *VaultAlbumsManager) PreCacheAlbums(key []byte) {
	data, err := am.ReadAlbums(key)

	if err != nil {
		LogError(err)
		return
	}

	// Pre-cache thumbnails
	for album_id := range data.Albums {
		am.thumbnail_cache.GetAlbumThumbnail(album_id, key)
	}

	LogDebug("Pre-cached albums")
}

// Resolves the path of a thumbnail asset file
// asset_id - Asset file ID
// Returns the path to the file
func (am *VaultAlbumsManager) GetThumbnailAssetPath(asset_id uint64) string {
	return path.Join(am.path, "thumb_album", "s_"+fmt.Sprint(asset_id)+".pma")
}

// Acquires a thumbnail asset file, creating a read/write lock for it
// asset_id - Asset file ID
// Returns:
//
//	1 - True if the asset was acquired, false if the asset was deleted
//	2 - The full path to the asset file
//	3 - The lock to control access to the file
func (am *VaultAlbumsManager) AcquireThumbnailAsset(asset_id uint64) (bool, string, *ReadWriteLock) {
	am.thumbnailsMu.Lock()
	defer am.thumbnailsMu.Unlock()

	p := am.GetThumbnailAssetPath(asset_id)

	if am.thumbnails[asset_id] != nil {
		am.thumbnails[asset_id].use_count++
		return true, p, am.thumbnails[asset_id].lock
	}

	f := AlbumThumbnailAsset{
		id:        asset_id,
		lock:      CreateReadWriteLock(),
		use_count: 1,
	}

	am.thumbnails[asset_id] = &f

	return true, p, f.lock
}

// Releases a thumbnail asset file
// Must be called to release the resources created by AcquireThumbnailAsset()
// asset_id - Asset file ID
func (am *VaultAlbumsManager) ReleaseThumbnailAsset(asset_id uint64) {
	am.thumbnailsMu.Lock()
	defer am.thumbnailsMu.Unlock()

	if am.thumbnails[asset_id] != nil {
		am.thumbnails[asset_id].use_count--

		if am.thumbnails[asset_id].use_count <= 0 {
			delete(am.thumbnails, asset_id)
		}
	}
}
