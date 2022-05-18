// Album manager

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type VaultAlbumData struct {
	Name string   `json:"name"`
	List []uint64 `json:"list"`
}

type VaultAlbumsData struct {
	NextId uint64                     `json:"next_id"`
	Albums map[uint64]*VaultAlbumData `json:"albums"`
}

func (data *VaultAlbumData) HasMedia(media_id uint64) bool {
	for i := 0; i < len(data.List); i++ {
		if data.List[i] == media_id {
			return true
		}
	}

	return false
}

type VaultAlbumsManager struct {
	path string

	albums_file string
	lock        *ReadWriteLock
}

func ResolveAlbumFilePath(base_path string, album_id uint64) string {
	return path.Join(base_path, "tags", "album_"+fmt.Sprint(album_id)+".pmv")
}

func (am *VaultAlbumsManager) Initialize(base_path string) {
	am.path = base_path
	am.albums_file = path.Join(base_path, "albums.pmv")
	am.lock = CreateReadWriteLock()
}

func (am *VaultAlbumsManager) readData(key []byte) (*VaultAlbumsData, error) {
	if _, err := os.Stat(am.albums_file); err == nil {
		// Load file
		b, err := ioutil.ReadFile(am.albums_file)

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

func (am *VaultAlbumsManager) ReadAlbums(key []byte) (*VaultAlbumsData, error) {
	am.lock.StartRead() // Request read
	defer am.lock.EndRead()

	return am.readData(key)
}

func (am *VaultAlbumsManager) StartWrite(key []byte) (*VaultAlbumsData, error) {
	am.lock.RequestWrite() // Request write

	return am.readData(key)
}

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

	tmpFile := GetTemporalFileName("pmv")

	err = ioutil.WriteFile(tmpFile, encData, FILE_PERMISSION)

	if err != nil {
		return err
	}

	// Save to original file
	am.lock.StartWrite()

	err = os.Rename(tmpFile, am.albums_file)

	return err
}

func (am *VaultAlbumsManager) CancelWrite() {
	am.lock.EndWrite()
}

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

func (am *VaultAlbumsManager) AddMediaToAlbum(album_id uint64, media_id uint64, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil || data.Albums[album_id].HasMedia(media_id) {
		am.CancelWrite()
		return false, nil // Not found
	}

	old_list := data.Albums[album_id].List
	new_list := append(old_list, media_id)

	data.Albums[album_id].List = new_list

	err = am.EndWrite(data, key)

	return true, err
}

func (am *VaultAlbumsManager) RemoveMediaFromAlbum(album_id uint64, media_id uint64, key []byte) (bool, error) {
	data, err := am.StartWrite(key)

	if err != nil {
		return false, err
	}

	if data.Albums[album_id] == nil || !data.Albums[album_id].HasMedia(media_id) {
		am.CancelWrite()
		return false, nil // Not found
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
