// User configuration

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

// Video resolution config
type UserConfigResolution struct {
	Width  int32 `json:"width"`  // Width (PX)
	Height int32 `json:"height"` // Height (PX)
	Fps    int32 `json:"fps"`    // Frames per second

}

// Check if resolution is the same or smaller than the original resolution
// width - Original width
// height - Original height
// fps - Original frames per second
func (res UserConfigResolution) Fits(width int32, height int32, fps int32) bool {
	return (res.Width < width) && (res.Height < height) && (res.Fps < fps || res.Fps <= 30)
}

// Picture resolution config
type UserConfigImageResolution struct {
	Width  int32 `json:"width"`  // Width (PX)
	Height int32 `json:"height"` // Height (PX)
}

// Check if resolution is the same or smaller than the original resolution
// width - Original width
// height - Original height
func (res UserConfigImageResolution) Fits(width int32, height int32) bool {
	return (res.Width < width) && (res.Height < height)
}

// User vault configuration data
type UserConfig struct {
	CustomTitle      string                      `json:"title"`             // Custom title
	CustomCSS        string                      `json:"css"`               // Custom CSS code
	MaxTasks         int32                       `json:"max_tasks"`         // Max number of tasks in parallel
	EncodingThreads  int32                       `json:"encoding_threads"`  // Max encoding threads for FFMPEG
	Resolutions      []UserConfigResolution      `json:"resolutions"`       // Resolutions to encode (for videos)
	ImageResolutions []UserConfigImageResolution `json:"image_resolutions"` // Resolutions to encode (For pictures)
}

// User configuration manager
type UserConfigManager struct {
	file string         // User config file
	lock *ReadWriteLock // Lock to control access to the file
}

// Initializes user config manager
// base_path - Vault path
func (uc *UserConfigManager) Initialize(base_path string) {
	uc.file = path.Join(base_path, "user_config.pmv")
	uc.lock = CreateReadWriteLock()
}

// Reads user config
// key - Vault decryption key
// Returns user config data
func (uc *UserConfigManager) Read(key []byte) (*UserConfig, error) {
	if _, err := os.Stat(uc.file); err == nil {
		// Load file
		b, err := os.ReadFile(uc.file)

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
		}

		return &mp, nil
	} else {
		return nil, err
	}
}

// Writes user configuration
// data - Data to write
// key - Vault encryption key
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

	tmpFile := GetTemporalFileName("pmv", true)

	err = os.WriteFile(tmpFile, encData, FILE_PERMISSION)

	if err != nil {
		return err
	}

	// Save to original file
	uc.lock.StartWrite()

	err = os.Rename(tmpFile, uc.file)

	return err
}
