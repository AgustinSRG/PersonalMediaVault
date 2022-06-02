// User configuration

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
)

type UserConfigResolution struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
	Fps    int32 `json:"fps"`
}

type UserConfigImageResolution struct {
	Width  int32 `json:"width"`
	Height int32 `json:"height"`
}

type UserConfigPinnedTag struct {
	Tag   string `json:"tag"`
	Title string `json:"title"`
}

type UserConfig struct {
	MaxTasks         int32                       `json:"max_tasks"`
	EncodingThreads  int32                       `json:"encoding_threads"`
	Resolutions      []UserConfigResolution      `json:"resolutions"`
	ImageResolutions []UserConfigImageResolution `json:"image_resolutions"`
	PinnedTags       []UserConfigPinnedTag       `json:"pinned_tags"`
}

type UserConfigManager struct {
	file string
	lock *ReadWriteLock
}

func (uc *UserConfigManager) Initialize(base_path string) {
	uc.file = path.Join(base_path, "user_config.pmv")
}

func (uc *UserConfigManager) Read(key []byte) (*UserConfig, error) {
	if _, err := os.Stat(uc.file); err == nil {
		// Load file
		b, err := ioutil.ReadFile(uc.file)

		if err != nil {
			return nil, err
		}

		bJSON, err := decryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp UserConfig

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// Default config

		mp := UserConfig{
			MaxTasks:         1,
			EncodingThreads:  0,
			Resolutions:      make([]UserConfigResolution, 0),
			ImageResolutions: make([]UserConfigImageResolution, 0),
			PinnedTags:       make([]UserConfigPinnedTag, 0),
		}

		return &mp, nil
	} else {
		return nil, err
	}
}

func (uc *UserConfigManager) Write(data *UserConfig, key []byte) error {
	uc.lock.RequestWrite() // Request write
	defer uc.lock.EndWrite()

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
	uc.lock.StartWrite()

	err = os.Rename(tmpFile, uc.file)

	return err
}

func (res UserConfigResolution) Fits(width int32, height int32, fps int32) bool {
	return (res.Width < width) && (res.Height < height) && (res.Fps < fps || res.Fps <= 30)
}

func (res UserConfigImageResolution) Fits(width int32, height int32) bool {
	return (res.Width < width) && (res.Height < height)
}
