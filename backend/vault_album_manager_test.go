// Tests for album manager

package main

import (
	"crypto/rand"
	"os"
	"testing"
)

func compareAlbumLists(list1 []uint64, list2 []uint64) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i := 0; i < len(list1); i++ {
		if list1[i] != list2[i] {
			return false
		}
	}

	return true
}

func TestVaultAlbumManager(t *testing.T) {
	test_path_base := "./temp"

	SetTempFilesPath(test_path_base)

	err := os.MkdirAll(test_path_base, FOLDER_PERMISSION)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	// Generate a random key

	key := make([]byte, 32)
	rand.Read(key)

	// Initialize new albums manager

	var tm VaultAlbumsManager

	tm.Initialize(test_path_base)

	// Tests start here
	//////////////////////////

	album_id, err := tm.CreateAlbum("Example album", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	added, err := tm.AddMediaToAlbum(album_id, 1, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !added {
		t.Errorf("Media was not added to the album")
	}

	added, err = tm.AddMediaToAlbum(album_id, 2, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !added {
		t.Errorf("Media was not added to the album")
	}

	added, err = tm.AddMediaToAlbum(album_id, 3, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !added {
		t.Errorf("Media was not added to the album")
	}

	albums, err := tm.ReadAlbums(key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if albums.Albums[album_id] == nil {
		t.Errorf("Album is not in the list")
	}

	if albums.Albums[album_id].Name != "Example album" {
		t.Errorf("Expected name = (%s), but found (%s)", "Example album", albums.Albums[album_id].Name)
	}

	if !compareAlbumLists(albums.Albums[album_id].List, []uint64{1, 2, 3}) {
		t.Errorf("Expected list = (%v), but found (%v)", []uint64{1, 2, 3}, albums.Albums[album_id].List)
	}

	removed, err := tm.RemoveMediaFromAlbum(album_id, 2, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !removed {
		t.Errorf("Media was not removed")
	}

	albums, err = tm.ReadAlbums(key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if albums.Albums[album_id] == nil {
		t.Errorf("Album is not in the list")
	}

	if albums.Albums[album_id].Name != "Example album" {
		t.Errorf("Expected name = (%s), but found (%s)", "Example album", albums.Albums[album_id].Name)
	}

	if !compareAlbumLists(albums.Albums[album_id].List, []uint64{1, 3}) {
		t.Errorf("Expected list = (%v), but found (%v)", []uint64{1, 3}, albums.Albums[album_id].List)
	}

	renamed, err := tm.RenameAlbum(album_id, "Name of Album 1", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !renamed {
		t.Errorf("Album was not renamed")
	}

	album2, err := tm.CreateAlbum("Album2", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	album3, err := tm.CreateAlbum("Album3", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	added, err = tm.AddMediaToAlbum(album2, 1, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !added {
		t.Errorf("Media was not added to the album")
	}

	added, err = tm.AddMediaToAlbum(album3, 2, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !added {
		t.Errorf("Media was not added to the album")
	}

	albums, err = tm.ReadAlbums(key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if albums.Albums[album_id] == nil {
		t.Errorf("Album is not in the list")
	}

	if albums.Albums[album_id].Name != "Name of Album 1" {
		t.Errorf("Expected name = (%s), but found (%s)", "Name of Album 1", albums.Albums[album_id].Name)
	}

	if !compareAlbumLists(albums.Albums[album_id].List, []uint64{1, 3}) {
		t.Errorf("Expected list = (%v), but found (%v)", []uint64{1, 3}, albums.Albums[album_id].List)
	}

	if albums.Albums[album2] == nil {
		t.Errorf("Album is not in the list")
	}

	if albums.Albums[album2].Name != "Album2" {
		t.Errorf("Expected name = (%s), but found (%s)", "Name of Album 1", albums.Albums[album2].Name)
	}

	if !compareAlbumLists(albums.Albums[album2].List, []uint64{1}) {
		t.Errorf("Expected list = (%v), but found (%v)", []uint64{1}, albums.Albums[album2].List)
	}

	if albums.Albums[album3] == nil {
		t.Errorf("Album is not in the list")
	}

	if albums.Albums[album3].Name != "Album3" {
		t.Errorf("Expected name = (%s), but found (%s)", "Name of Album 1", albums.Albums[album3].Name)
	}

	if !compareAlbumLists(albums.Albums[album3].List, []uint64{2}) {
		t.Errorf("Expected list = (%v), but found (%v)", []uint64{2}, albums.Albums[album3].List)
	}

	//////////////////////////
	// Tests end here

	ClearTemporalFilesPath() // Remove all files
}
