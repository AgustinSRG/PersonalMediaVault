// Tests for tag manager

package main

import (
	"crypto/rand"
	"os"
	"testing"
)

func TestVaultTagManager(t *testing.T) {
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

	// Initialize new tag manager

	var tm VaultTagManager

	err = tm.Initialize(test_path_base)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	// Tests start here
	//////////////////////////

	err = tm.TagMedia(1, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = tm.TagMedia(2, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = tm.TagMedia(3, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	list, err := tm.ListTaggedMedia("tag_example", key, 0, 100, false)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if len(list) != 3 || list[0] != 1 || list[1] != 2 || list[2] != 3 {
		t.Errorf("Invalid list: %v", list)
	}

	check, err := tm.CheckMediaTag(1, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !check {
		t.Errorf("Expected tagged media, but it wasn't tagged")
	}

	check, err = tm.CheckMediaTag(2, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !check {
		t.Errorf("Expected tagged media, but it wasn't tagged")
	}

	check, err = tm.CheckMediaTag(3, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if !check {
		t.Errorf("Expected tagged media, but it wasn't tagged")
	}

	check, err = tm.CheckMediaTag(5, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if check {
		t.Errorf("Expected untagged media, but it was tagged")
	}

	err = tm.TagMedia(6, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = tm.TagMedia(6, "tag_2", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	tagList, err := tm.ReadList(key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	found, tag_example := tagList.FindTag("tag_example")

	if !found {
		t.Errorf("Tag not found")
	}

	found, _ = tagList.FindTag("tag_2")

	if !found {
		t.Errorf("Tag not found")
	}

	err = tm.UnTagMedia(2, tag_example, key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	check, err = tm.CheckMediaTag(2, "tag_example", key)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if check {
		t.Errorf("Expected untagged media, but it was tagged")
	}

	list, err = tm.ListTaggedMedia("tag_example", key, 0, 100, false)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if len(list) != 3 || list[0] != 1 || list[1] != 3 || list[2] != 6 {
		t.Errorf("Invalid list: %v", list)
	}

	list, err = tm.ListTaggedMedia("tag_2", key, 0, 100, false)

	if err != nil {
		t.Error(err)
		panic(err)
	}

	if len(list) != 1 || list[0] != 6 {
		t.Errorf("Invalid list: %v", list)
	}

	//////////////////////////
	// Tests end here

	ClearTemporalFilesPath() // Remove all files
}