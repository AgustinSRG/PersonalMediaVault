// Album manager

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

// Album data
type VaultAlbumData struct {
	Name string   `json:"name"` // Name of the album
	List []uint64 `json:"list"` // Ordered list of media to play
}

// Album list data
type VaultAlbumsData struct {
	NextId uint64                     `json:"next_id"` // Id for the next album to create
	Albums map[uint64]*VaultAlbumData `json:"albums"`  // Albums (Id -> Data)
}

// Checks if an album has a media in it
// media_id - Id of the media file
// Return strue if the album has the media in its list
func (data *VaultAlbumData) HasMedia(media_id uint64) bool {
	for i := 0; i < len(data.List); i++ {
		if data.List[i] == media_id {
			return true
		}
	}

	return false
}

// Album manager
type VaultAlbumsManager struct {
	path string // Vault path

	albums_file string         // Path to the albums data file
	lock        *ReadWriteLock // Lock to control access to the file
}

// Initializes albums manager
// base_path - Vault path
func (am *VaultAlbumsManager) Initialize(base_path string) {
	am.path = base_path
	am.albums_file = path.Join(base_path, "albums.pmv")
	am.lock = CreateReadWriteLock()
}

// Reads albums list data
// key - Vault decryption key
// Returns the data
// Thread unsafe: This is an internal method
func (am *VaultAlbumsManager) readData(key []byte) (*VaultAlbumsData, error) {
	if _, err := os.Stat(am.albums_file); err == nil {
		// Load file
		b, err := os.ReadFile(am.albums_file)

		if err != nil {
			return nil, err
		}

		bJSON, err := decryptFileContents(b, key)

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

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No albums yet

		mp := VaultAlbumsData{
			NextId: 0,
			Albums: make(map[uint64]*VaultAlbumData),
		}

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

	return am.readData(key)
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

	encData, err := encryptFileContents(jsonData, AES256_ZIP, key)

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

	err = os.Rename(tmpFile, am.albums_file)

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
		Name: name,
		List: make([]uint64, 0),
	}

	err = am.EndWrite(data, key)

	return album_id, err
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
	new_list := append(old_list, media_id)

	data.Albums[album_id].List = new_list

	err = am.EndWrite(data, key)

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
	new_list := make([]uint64, 0)

	for i := 0; i < len(old_list); i++ {
		if old_list[i] != media_id {
			new_list = append(new_list, old_list[i])
		}
	}

	data.Albums[album_id].List = new_list

	err = am.EndWrite(data, key)

	return true, err
}

// Sets album media list
// album_id - Album ID
// media_list - List of media files
// key - Vault encryption key
// Returns true if sucess
func (am *VaultAlbumsManager) SetAlbumList(album_id uint64, media_list []uint64, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil {
		am.CancelWrite()
		return false, nil // Not found
	}

	data.Albums[album_id].List = media_list

	err = am.EndWrite(data, key)

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

	err = am.EndWrite(data, key)

	return true, err
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
	}

	err = am.EndWrite(data, key)

	return err
}
