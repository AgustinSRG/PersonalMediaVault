// Media asset

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

// Media types

type MediaType uint16

const (
	MediaTypeDeleted MediaType = 0 // Reserved for not found
	MediaTypeImage   MediaType = 1 // Image (png)
	MediaTypeVideo   MediaType = 2 // Video (mp4)
	MediaTypeAudio   MediaType = 3 // Audio (mp3)
)

// Represents a media asset stored in disk
type MediaAsset struct {
	id uint64 // Media ID

	path string         // Path of the media asset in disk
	lock *ReadWriteLock // Lock to control read/write operations

	use_count int32 // Counter of threads accessing the asset

	deleting bool // True if the asset is being deleted

	mu *sync.Mutex // Mutex to control multi-threading access to the struct

	files map[uint64]*MediaAssetFile // Files that the media asset has
}

// Represents an asset file that stores data of a media asset
type MediaAssetFile struct {
	id uint64 // File ID

	lock *ReadWriteLock // Lock to control read/write operations

	use_count int32 // Counter of threads accessing the file

	waiting    bool            // True if the is a thread waiting for the file to be released
	wait_group *sync.WaitGroup // Wait group to wait for the asset to be released
}

// Contains data of a resolution for the media asset
type MediaResolution struct {
	Width  int32 `json:"width"`  // Width (px)
	Height int32 `json:"height"` // Height (px)
	Fps    int32 `json:"fps"`    // Frames per second

	Ready     bool   `json:"ready"` // True if the resolution is ready for playback
	Asset     uint64 `json:"asset"` // ID of the asset file (MediaAssetFile) where the media is stored
	Extension string `json:"ext"`   // Extension of the media file

	TaskId uint64 `json:"task_id"` // Id of the task that must encode the resolution (only when Ready = false)
}

// Contains data of a subtitles file
type MediaSubtitle struct {
	Id    string `json:"id"`    // ID of the subtitles (language ISO)
	Name  string `json:"name"`  // Name of the subtitles for user display
	Asset uint64 `json:"asset"` // ID of the asset file (MediaAssetFile) where the subtitles are stored
}

// Contains data of a time split (for videos/audios)
type MediaSplit struct {
	Time float64 `json:"time"` // Time in seconds of the split
	Name string  `json:"name"` // Name of the chapter that starts at this time
}

// Contains data of an audio track
type MediaAudioTrack struct {
	Id    string `json:"id"`    // ID of the track (language ISO)
	Name  string `json:"name"`  // Name of the track
	Asset uint64 `json:"asset"` // ID of the asset file (MediaAssetFile) where the audio is stored
}

// Contains data of an attachment
type MediaAttachment struct {
	Name  string `json:"name"`  // Name of the attachment
	Asset uint64 `json:"asset"` // ID of the asset file (MediaAssetFile) where the attachment is stored
	Size  uint64 `json:"size"`  // Size in bytes
}

// Contains the metadata of a media asset
type MediaMetadata struct {
	Id uint64 `json:"id"` // Media ID

	Type MediaType `json:"type"` // Type of media (image, video, audio)

	Title       string   `json:"title"`                 // Title
	Description string   `json:"description,omitempty"` // Description
	Tags        []uint64 `json:"tags"`                  // List of tag IDs

	MediaDuration float64 `json:"duration,omitempty"` // Duration (seconds)
	Width         int32   `json:"width,omitempty"`    // Width (px)
	Height        int32   `json:"height,omitempty"`   // Height (px)
	Fps           int32   `json:"fps,omitempty"`      // Frames per second

	UploadTimestamp int64 `json:"upload_time"` // Upload timestamp (unix milliseconds)

	NextAssetID uint64 `json:"next_asset_id"` // Id to give to the next asset file

	OriginalReady     bool   `json:"original_ready"`          // True if original media asset is fully uploaded and encrypted
	OriginalAsset     uint64 `json:"original_asset"`          // ID of the asset file (MediaAssetFile) where the media is stored
	OriginalExtension string `json:"original_ext"`            // Extension of the original media file
	OriginalTask      uint64 `json:"original_task,omitempty"` // ID of the task that must encode the original asset (only if OriginalEncoded = false)
	OriginalEncoded   bool   `json:"original_encoded"`        // True if the original asset is ready for playback

	ThumbnailReady bool   `json:"thumb_ready"` // True if the thumbnail is ready to be displayed
	ThumbnailAsset uint64 `json:"thumb_asset"` // ID of the asset file (MediaAssetFile) where the thumbnail is stored

	Resolutions []MediaResolution `json:"resolutions,omitempty"`  // List of extra resolutions (not original)
	Subtitles   []MediaSubtitle   `json:"subtitles,omitempty"`    // List of subtitle files
	Splits      []MediaSplit      `json:"time_splits,omitempty"`  // List of time splits
	AudioTracks []MediaAudioTrack `json:"audio_tracks,omitempty"` // List of audio tracks
	Attachments []MediaAttachment `json:"attachments,omitempty"`  // List of attachments

	PreviewsReady    bool    `json:"previews_ready,omitempty"`    // True if timeline previews are ready to be displayed
	PreviewsTask     uint64  `json:"previews_task,omitempty"`     // ID of the task that must create the video previews (only if PreviewsReady = false)
	PreviewsInterval float64 `json:"previews_interval,omitempty"` // Interval for each video preview, in seconds
	PreviewsAsset    uint64  `json:"previews_asset,omitempty"`    // ID of the asset file (MediaAssetFile) where the video previews are stored

	ForceStartBeginning bool `json:"force_start_beginning,omitempty"` // True to indicate clients not to save the current time for this media
	IsAnimation         bool `json:"is_anim,omitempty"`               // True if the media is an animation

	HasImageNotes   bool   `json:"img_notes,omitempty"`       // True to indicate the asset has image notes
	ImageNotesAsset uint64 `json:"img_notes_asset,omitempty"` // Asset where the image notes are stored

	HasExtendedDescription   bool   `json:"ext_desc,omitempty"`       // True to indicate the asset has extended description
	ExtendedDescriptionAsset uint64 `json:"ext_desc_asset,omitempty"` // Asset where the extended description is stored
}

// Creates a new media asset. Creates the folder and stores the initial metadata.
// Used in the upload process
// key - Vault encryption key
// media_type - Type of media
// title - Title
// desc - Description
// duration - Duration in seconds
// width - Width (px)
// height - Height (px)
// fps - Frames per second
func (media *MediaAsset) CreateNewMediaAsset(key []byte, media_type MediaType, title string, desc string, duration float64, width int32, height int32, fps int32) error {
	// Create the folder
	err := os.MkdirAll(media.path, FOLDER_PERMISSION)

	if err != nil {
		LogError(err)
	}

	now := time.Now().UnixMilli()

	isShortAnimation := media_type == MediaTypeVideo && duration < 10

	meta := MediaMetadata{
		Id:                       media.id,
		Type:                     media_type,
		MediaDuration:            duration,
		Width:                    width,
		Height:                   height,
		Fps:                      fps,
		Title:                    title,
		Description:              desc,
		Tags:                     make([]uint64, 0),
		UploadTimestamp:          now,
		NextAssetID:              0,
		OriginalReady:            false,
		OriginalAsset:            0,
		OriginalTask:             0,
		OriginalEncoded:          false,
		OriginalExtension:        "",
		ThumbnailReady:           false,
		ThumbnailAsset:           0,
		Resolutions:              make([]MediaResolution, 0),
		Subtitles:                make([]MediaSubtitle, 0),
		AudioTracks:              make([]MediaAudioTrack, 0),
		Splits:                   make([]MediaSplit, 0),
		Attachments:              make([]MediaAttachment, 0),
		PreviewsReady:            false,
		PreviewsInterval:         0,
		PreviewsAsset:            0,
		HasImageNotes:            false,
		ImageNotesAsset:          0,
		HasExtendedDescription:   false,
		ExtendedDescriptionAsset: 0,
		IsAnimation:              isShortAnimation,
		ForceStartBeginning:      isShortAnimation,
	}

	media.lock.RequestWrite() // Request write
	defer media.lock.EndWrite()

	jsonData, err := json.Marshal(meta)

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
	media.lock.StartWrite()

	err = RenameAndReplace(tmpFile, path.Join(media.path, "meta.pmv"))

	return err
}

// Reads metadata
// key - Vault decryption key
// Note: Internal method, do not call this from outside
func (media *MediaAsset) readData(key []byte) (*MediaMetadata, error) {
	file := path.Join(media.path, "meta.pmv")
	if _, err := os.Stat(file); err == nil {
		// Load file
		b, err := os.ReadFile(file)

		if err != nil {
			return nil, err
		}

		bJSON, err := encrypted_storage.DecryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp MediaMetadata

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		if mp.Resolutions == nil {
			mp.Resolutions = make([]MediaResolution, 0)
		}

		if mp.Subtitles == nil {
			mp.Subtitles = make([]MediaSubtitle, 0)
		}

		if mp.AudioTracks == nil {
			mp.AudioTracks = make([]MediaAudioTrack, 0)
		}

		if mp.Splits == nil {
			mp.Splits = make([]MediaSplit, 0)
		}

		if mp.Attachments == nil {
			mp.Attachments = make([]MediaAttachment, 0)
		}

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No tags

		return nil, nil
	} else {
		return nil, err
	}
}

// Reads metadata
// key - Vault decryption key
func (media *MediaAsset) ReadMetadata(key []byte) (*MediaMetadata, error) {
	media.lock.StartRead() // Request read
	defer media.lock.EndRead()

	return media.readData(key)
}

// Gets the size of the metadata file (in bytes)
func (media *MediaAsset) GetMetadataSize() (int64, error) {
	media.lock.StartRead() // Request read
	defer media.lock.EndRead()

	file := path.Join(media.path, "meta.pmv")

	stats, err := os.Stat(file)

	if err != nil {
		return 0, err
	}

	return stats.Size(), nil
}

// Starts a write operation and reads the metadata, returning it
// key - Vault decryption key
func (media *MediaAsset) StartWrite(key []byte) (*MediaMetadata, error) {
	media.lock.RequestWrite() // Request write

	return media.readData(key)
}

// Starts a write operation and reads the metadata, returning it
// Also: This method waits for any threads to finish, and ensures the resource is locked until you release it
// key - Vault decryption key
func (media *MediaAsset) StartWriteWithFullLock(key []byte) (*MediaMetadata, error) {
	media.lock.RequestWrite() // Request write
	media.lock.StartWrite()

	return media.readData(key)
}

// Finish a write operation on the metadata
// data - New metadata to write
// key - Vault encryption key
// hasFullLock - True if you used StartWriteWithFullLock() to start the operation
// This method also unlocks the resource
func (media *MediaAsset) EndWrite(data *MediaMetadata, key []byte, hasFullLock bool) error {
	defer media.lock.EndWrite()

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
	if !hasFullLock {
		media.lock.StartWrite()
	}

	err = RenameAndReplace(tmpFile, path.Join(media.path, "meta.pmv"))

	// LogDebug("WRITE META: " + string(jsonData))

	return err
}

// Cancels a write operation
// Unlocks the resource
func (media *MediaAsset) CancelWrite() {
	media.lock.EndWrite()
}

// Types of media asset files
const (
	ASSET_MULTI_FILE  = "m" // File containing multiple encrypted files (similar to tar)
	ASSET_SINGLE_FILE = "s" // File containing a single chunked encrypted file
)

// Resolves the path of an asset file
// asset_id - Asset file ID
// asset_type - Asset type (ASSET_MULTI_FILE or ASSET_SINGLE_FILE)
// Returns the path to the file
func (media *MediaAsset) GetAssetPath(asset_id uint64, asset_type string) string {
	return path.Join(media.path, asset_type+"_"+fmt.Sprint(asset_id)+".pma")
}

// Acquires an asset file, creating a read/write lock for it
// asset_id - Asset file ID
// asset_type - Asset type (ASSET_MULTI_FILE or ASSET_SINGLE_FILE)
// Returns:
//
//	1 - True if the asset was acquired, false if the asset was deleted
//	2 - The full path to the asset file
//	3 - The lock to control access to the file
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
		id:         asset_id,
		lock:       CreateReadWriteLock(),
		use_count:  1,
		waiting:    false,
		wait_group: nil,
	}

	media.files[asset_id] = &f

	return true, p, f.lock
}

// Releases an asset file
// Must be called to release the resources created by AcquireAsset()
// asset_id - Asset file ID
func (media *MediaAsset) ReleaseAsset(asset_id uint64) {
	media.mu.Lock()
	defer media.mu.Unlock()

	if media.files[asset_id] != nil {
		media.files[asset_id].use_count--

		if media.files[asset_id].use_count <= 0 {

			if media.files[asset_id].waiting {
				if media.files[asset_id].wait_group != nil {
					media.files[asset_id].wait_group.Done()
					media.files[asset_id].wait_group = nil
				}
			}

			delete(media.files, asset_id)
		}
	}
}

// Deletes the media asset
// Deletes every single file
func (media *MediaAsset) Delete() {
	// Delete metadata file

	media.lock.RequestWrite()
	media.lock.StartWrite()

	os.Remove(path.Join(media.path, "meta.pmv"))

	media.lock.EndWrite()

	go media.deleteAll()
}

// Deletes every file contained in the media asset
func (media *MediaAsset) deleteAll() {
	// Set deleting and wait for assets to be released

	waitGroups := make([]*sync.WaitGroup, 0)

	media.mu.Lock()

	if media.deleting {
		return
	}

	media.deleting = true

	for _, a := range media.files {
		a.waiting = true
		if a.wait_group == nil {
			a.wait_group = &sync.WaitGroup{}
			a.wait_group.Add(1)
		}
		waitGroups = append(waitGroups, a.wait_group)
	}

	media.mu.Unlock()

	for i := 0; i < len(waitGroups); i++ {
		waitGroups[i].Wait()
	}

	// Now, delete everything

	media.lock.RequestWrite()
	media.lock.StartWrite()

	os.RemoveAll(media.path)

	media.lock.EndWrite()
}

// Adds a tag to the list
// tag - tag ID
func (meta *MediaMetadata) AddTag(tag uint64) {
	found := false

	for i := 0; i < len(meta.Tags); i++ {
		if meta.Tags[i] == tag {
			found = true
			break
		}
	}

	if !found {
		meta.Tags = append(meta.Tags, tag)
	}
}

// Removes a tag from the list
// tag - tag ID
func (meta *MediaMetadata) RemoveTag(tag uint64) {
	newTags := make([]uint64, 0)

	for i := 0; i < len(meta.Tags); i++ {
		if meta.Tags[i] != tag {
			newTags = append(newTags, meta.Tags[i])
		}
	}

	meta.Tags = newTags
}

// Finds a resolution
// width - Width (px)
// height - Height (px)
// fps - Frames per second
// Returns the index in the array (if found), or -1 if not found
func (meta *MediaMetadata) FindResolution(width int32, height int32, fps int32) int {
	for i := 0; i < len(meta.Resolutions); i++ {
		if meta.Resolutions[i].Width == width && meta.Resolutions[i].Height == height && meta.Resolutions[i].Fps == fps {
			return i
		}
	}

	return -1
}

// Removes a resolution given the index
// index - Resolution index in the array
func (meta *MediaMetadata) RemoveResolution(index int) {
	newResolutions := make([]MediaResolution, 0)

	for i := 0; i < len(meta.Resolutions); i++ {
		if i != index {
			newResolutions = append(newResolutions, meta.Resolutions[i])
		}
	}

	meta.Resolutions = newResolutions
}

// Finds a subtitle in the subtitles array
// id - Subtitles ID (language ISO)
func (meta *MediaMetadata) FindSubtitle(id string) int {
	for i := 0; i < len(meta.Subtitles); i++ {
		if meta.Subtitles[i].Id == id {
			return i
		}
	}

	return -1
}

// Removes a subtitle given the index
// index - Subtitles index in the array
func (meta *MediaMetadata) RemoveSubtitle(index int) {
	newSubtitles := make([]MediaSubtitle, 0)

	for i := 0; i < len(meta.Subtitles); i++ {
		if i != index {
			newSubtitles = append(newSubtitles, meta.Subtitles[i])
		}
	}

	meta.Subtitles = newSubtitles
}

// Adds a subtitle entry
// id - Subtitle ID
// name - Display name
// asset - Asset file ID containing the subtitles
func (meta *MediaMetadata) AddSubtitle(id string, name string, asset uint64) {
	sub := MediaSubtitle{
		Id:    id,
		Name:  name,
		Asset: asset,
	}
	meta.Subtitles = append(meta.Subtitles, sub)
}

// Finds an audio track in the audios array
// id - Track ID (language ISO)
func (meta *MediaMetadata) FindAudioTrack(id string) int {
	for i := 0; i < len(meta.AudioTracks); i++ {
		if meta.AudioTracks[i].Id == id {
			return i
		}
	}

	return -1
}

// Removes an audio track given the index
// index - Track index in the array
func (meta *MediaMetadata) RemoveAudioTrack(index int) {
	newTracks := make([]MediaAudioTrack, 0)

	for i := 0; i < len(meta.AudioTracks); i++ {
		if i != index {
			newTracks = append(newTracks, meta.AudioTracks[i])
		}
	}

	meta.AudioTracks = newTracks
}

// Adds a track entry
// id - Track ID
// name - Display name
// asset - Asset file ID containing the track
func (meta *MediaMetadata) AddAudioTrack(id string, name string, asset uint64) {
	track := MediaAudioTrack{
		Id:    id,
		Name:  name,
		Asset: asset,
	}
	meta.AudioTracks = append(meta.AudioTracks, track)
}

// Adds an attachment
// name - Display name
// asset - Asset file ID containing the attachment
// size - Size in bytes of the attachment
func (meta *MediaMetadata) AddAttachment(name string, asset uint64, size uint64) {
	attachment := MediaAttachment{
		Name:  name,
		Asset: asset,
		Size:  size,
	}
	meta.Attachments = append(meta.Attachments, attachment)
}

// Finds an attachment in the array
// id - Asset ID
func (meta *MediaMetadata) FindAttachment(id uint64) int {
	for i := 0; i < len(meta.Attachments); i++ {
		if meta.Attachments[i].Asset == id {
			return i
		}
	}

	return -1
}

// Removes an attachment given the index
// index - Attachment index in the array
func (meta *MediaMetadata) RemoveAttachment(index int) {
	newAttachment := make([]MediaAttachment, 0)

	for i := 0; i < len(meta.Attachments); i++ {
		if i != index {
			newAttachment = append(newAttachment, meta.Attachments[i])
		}
	}

	meta.Attachments = newAttachment
}
