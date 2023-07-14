// Vault tags manager

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

// Tag manager
type VaultTagManager struct {
	path string // Vault path

	tag_list_file string         // Path to the tag list file
	tag_list_lock *ReadWriteLock // Lock to control acess to the file

	indexes map[uint64]*VaultTagIndexEntry // Indexes for each tag

	lock *sync.Mutex // Lock to control acess to the indexes map
}

// Tag Index manager data struct
type VaultTagIndexEntry struct {
	count        int64           // Number of threads accessing the index
	index        *VaultMainIndex // Index
	check_delete bool            // True if the tag is being deleted
}

// Data inside the tag list file
type VaultTagListData struct {
	NextId uint64            `json:"next_id"` // ID for the next tag to add
	Tags   map[uint64]string `json:"tags"`    // Tags Map (id -> name)
}

// Finds a tag by name
// tag_name - Name of the tag
// Returns true if it was found, false if if was not found
// Also returns the ID of the found tag
func (data *VaultTagListData) FindTag(tag_name string) (bool, uint64) {
	parsedName := ParseTagName(tag_name)
	for key, val := range data.Tags {
		if val == parsedName {
			return true, key
		}
	}
	return false, 0
}

// Parses tag name
// Removes line breaks and other weird stuff from the name
// Also converts to lowercase
func ParseTagName(name string) string {
	return strings.ToLower(
		strings.ReplaceAll(
			strings.Trim(
				strings.ReplaceAll(
					strings.ReplaceAll(name, "\n", " "),
					"\r",
					"",
				),
				" ",
			),
			" ",
			"_",
		),
	)
}

// Given a tag ID generates the path to the index file
// base_path - Vault path
// tag_id - Tag ID
// Returns the path to the index file
func ResolveTagIndexFilePath(base_path string, tag_id uint64) string {
	return path.Join(base_path, "tags", "tag_"+fmt.Sprint(tag_id)+".index")
}

// Initailizes the tag manager
// base_path - Vault path
func (tm *VaultTagManager) Initialize(base_path string) error {
	tm.path = base_path
	tm.tag_list_file = path.Join(base_path, "tag_list.pmv")
	tm.tag_list_lock = CreateReadWriteLock()
	tm.lock = &sync.Mutex{}
	tm.indexes = make(map[uint64]*VaultTagIndexEntry)

	// Make tags folder if not present
	err := os.MkdirAll(path.Join(base_path, "tags"), FOLDER_PERMISSION)

	if err != nil {
		return err
	}

	return nil
}

// Reads data from the tag list file
// key - Vault decryption key
// Returns the tag list data
// Warning: Thread unsafe. do not call this method externally, this is used internally
func (tm *VaultTagManager) readData(key []byte) (*VaultTagListData, error) {
	if _, err := os.Stat(tm.tag_list_file); err == nil {
		// Load file
		b, err := os.ReadFile(tm.tag_list_file)

		if err != nil {
			return nil, err
		}

		bJSON, err := encrypted_storage.DecryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp VaultTagListData

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// No tags

		mp := VaultTagListData{
			NextId: 0,
			Tags:   make(map[uint64]string),
		}

		return &mp, nil
	} else {
		return nil, err
	}
}

// Reads data from the tag list file
// key - Vault decryption key
// Returns the tag list data
func (tm *VaultTagManager) ReadList(key []byte) (*VaultTagListData, error) {
	tm.tag_list_lock.StartRead() // Request read
	defer tm.tag_list_lock.EndRead()

	return tm.readData(key)
}

// Starts a modification on the tag list file
// key - Vault decryption key
// Returns the tag list data
// After calling, the file is locked until you call EndWrite() or CancelWrite()
func (tm *VaultTagManager) StartWrite(key []byte) (*VaultTagListData, error) {
	tm.tag_list_lock.RequestWrite() // Request write

	return tm.readData(key)
}

// Applies a modification in the tag list file
// data - Data to write
// key - Vault encryption key
func (tm *VaultTagManager) EndWrite(data *VaultTagListData, key []byte) error {
	defer tm.tag_list_lock.EndWrite()

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
	tm.tag_list_lock.StartWrite()

	err = RenameAndReplace(tmpFile, tm.tag_list_file)

	return err
}

// Cancels a modification in the tag list file
func (tm *VaultTagManager) CancelWrite() {
	tm.tag_list_lock.EndWrite()
}

// Acquires an index file for a tag
// tag_id - Id for the tag
// Returns a reference to the index
// After this methos is called, an entry is reserved in memory.
// To release it, make sure to call ReleaseIndexFile()
func (tm *VaultTagManager) AcquireIndexFile(tag_id uint64) (*VaultMainIndex, error) {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if tm.indexes[tag_id] != nil {
		tm.indexes[tag_id].count++
		return tm.indexes[tag_id].index, nil
	}

	newIndex := VaultMainIndex{}
	err := newIndex.Initialize(ResolveTagIndexFilePath(tm.path, tag_id))

	if err != nil {
		return nil, err
	}

	tm.indexes[tag_id] = &VaultTagIndexEntry{
		count:        1,
		index:        &newIndex,
		check_delete: false,
	}

	return &newIndex, nil
}

// Releases an entry for a tag index file
// tag_id - Id of the tag
// check_delete - True if the modification made deleted the file
// key - Vault encryption key
func (tm *VaultTagManager) ReleaseIndexFile(tag_id uint64, check_delete bool, key []byte) {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if tm.indexes[tag_id] != nil {
		tm.indexes[tag_id].count--

		if check_delete {
			tm.indexes[tag_id].check_delete = true
		}

		if tm.indexes[tag_id].count <= 0 {
			if tm.indexes[tag_id].check_delete {
				tm.checkTagIndexToRemove(tag_id, tm.indexes[tag_id].index, key)
			}

			delete(tm.indexes, tag_id)
		}
	}
}

// Internal method, called to check if a tag index file requires deletion
// tag_id - Id of the tag
// index - Reference to the index
// key - Vault encryption key
func (tm *VaultTagManager) checkTagIndexToRemove(tag_id uint64, index *VaultMainIndex, key []byte) {
	r, err := index.StartRead()

	if err != nil {
		LogError(err)
		return
	}

	c, err := r.Count()

	if err != nil {
		index.EndRead(r)
		LogError(err)
		return
	}

	index.EndRead(r)

	if c == 0 {
		// Is empty, remove tag
		err = tm.RemoveTagFromList(tag_id, key)

		if err != nil {
			LogError(err)
			return
		}

		// Remove file
		err = index.Delete()

		if err != nil {
			LogError(err)
		}
	}
}

// Adds a tag
// tag_name - Tag name
// key - vault encryption key
// Returns the ID for the tag
func (tm *VaultTagManager) AddTagToList(tag_name string, key []byte) (uint64, error) {
	parsedTagName := ParseTagName(tag_name)

	data, err := tm.StartWrite(key)

	if err != nil {
		return 0, err
	}

	found, tag_id := data.FindTag(parsedTagName)

	if found {
		tm.CancelWrite()
		return tag_id, nil // Already exists
	}

	newIdForTag := data.NextId
	data.NextId++

	data.Tags[newIdForTag] = parsedTagName

	err = tm.EndWrite(data, key)

	return newIdForTag, err
}

// Removes tag from the list
// tag_id - Tag ID
// key - Vault encryption key
func (tm *VaultTagManager) RemoveTagFromList(tag_id uint64, key []byte) error {
	data, err := tm.StartWrite(key)

	if err != nil {
		return err
	}

	if data.Tags[tag_id] == "" {
		// Nothing to do
		tm.CancelWrite()
		return nil
	}

	delete(data.Tags, tag_id)

	err = tm.EndWrite(data, key)

	return err
}

// Adds a tag to a media file
// media_id - ID of the media file
// tag_name - Name of the tag
// key - Vault encryption key
// Returns the ID for the tag
func (tm *VaultTagManager) TagMedia(media_id uint64, tag_name string, key []byte) (uint64, error) {
	tag_id, err := tm.AddTagToList(tag_name, key)

	if err != nil {
		return 0, err
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return 0, err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	r, err := indexFile.StartWrite()

	if err != nil {
		return 0, err
	}

	added, _, err := r.file.AddValue(media_id)

	if err != nil {
		indexFile.CancelWrite(r)
		return 0, err
	}

	if !added {
		// Already tagged
		indexFile.CancelWrite(r)
		return tag_id, nil
	}

	err = indexFile.EndWrite(r)

	return tag_id, err
}

// Removes tag from a media file
// media_id - ID of the media file
// tag_name - Name of the tag
// key - Vault encryption key
func (tm *VaultTagManager) UnTagMedia(media_id uint64, tag_id uint64, key []byte) error {
	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return err
	}

	defer tm.ReleaseIndexFile(tag_id, true, key)

	r, err := indexFile.StartWrite()

	if err != nil {
		return err
	}

	removed, _, err := r.file.RemoveValue(media_id)

	if err != nil {
		indexFile.CancelWrite(r)
		return err
	}

	if !removed {
		// Was not tagged, no change
		indexFile.CancelWrite(r)
		return nil
	}

	err = indexFile.EndWrite(r)

	return err
}

// Checks if a media file is tagged by a specific tag
// media_id - ID of the media file
// tag_name - Name of the tag
// key - Vault encryption key
// Returns true if the media is tagged by the given tag
func (tm *VaultTagManager) CheckMediaTag(media_id uint64, tag_name string, key []byte) (bool, error) {
	tagList, err := tm.ReadList(key)

	if err != nil {
		return false, err
	}

	found, tag_id := tagList.FindTag(tag_name)

	if !found {
		return false, nil
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return false, err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	f, err := indexFile.StartRead()

	if err != nil {
		return false, err
	}

	found, _, err = f.BinarySearch(media_id)

	if err != nil {
		indexFile.EndRead(f)
		return false, err
	}

	indexFile.EndRead(f)

	return found, nil
}

// Gets a paginated list of all media tagged by a specific tag
// tag_name - Name of the tag
// key - Vault decryption key
// skip - Number of items to skip (for pagination)
// limit - Max number of items to return (for pagination)
// reverse - True for reverse order
// Returns (1) The list of media files (identifiers) tagged
// Returns (2) The total amount of items in the full list
// Returns (3) The tag ID
func (tm *VaultTagManager) ListTaggedMedia(tag_name string, key []byte, skip int64, limit int64, reverse bool) ([]uint64, int64, uint64, error) {
	tagList, err := tm.ReadList(key)

	if err != nil {
		return nil, 0, 0, err
	}

	found, tag_id := tagList.FindTag(tag_name)

	if !found {
		return make([]uint64, 0), 0, 0, nil
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return nil, 0, 0, err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	f, err := indexFile.StartRead()

	if err != nil {
		return nil, 0, 0, err
	}

	count, err := f.Count()

	if err != nil {
		indexFile.EndRead(f)
		return nil, 0, 0, err
	}

	var values []uint64

	if reverse {
		values, err = f.ListValuesReverse(skip, limit)

		if err != nil {
			indexFile.EndRead(f)
			return nil, 0, 0, err
		}
	} else {
		values, err = f.ListValues(skip, limit)

		if err != nil {
			indexFile.EndRead(f)
			return nil, 0, 0, err
		}
	}

	indexFile.EndRead(f)

	return values, count, tag_id, nil
}

// Returns a random set of media tagged by a tag
// tag_name - Name of the tag
// key - Vault decryption key
// seed - Seed for the random number generator
// limit - Number of items to return
// Returns the list of media files (identifiers) and the tag ID
func (tm *VaultTagManager) RandomTaggedMedia(tag_name string, key []byte, seed int64, limit int64) ([]uint64, uint64, error) {
	tagList, err := tm.ReadList(key)

	if err != nil {
		return nil, 0, err
	}

	found, tag_id := tagList.FindTag(tag_name)

	if !found {
		return make([]uint64, 0), 0, nil
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return nil, 0, err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	f, err := indexFile.StartRead()

	if err != nil {
		return nil, 0, err
	}

	var values []uint64

	values, err = f.RandomValues(seed, limit)

	if err != nil {
		indexFile.EndRead(f)
		return nil, 0, err
	}

	indexFile.EndRead(f)

	return values, tag_id, nil
}
