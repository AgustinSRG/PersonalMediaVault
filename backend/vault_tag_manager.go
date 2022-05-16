// Vault tags manager

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

type VaultTagManager struct {
	path string

	tag_list_file string
	tag_list_lock *ReadWriteLock

	lock *sync.Mutex

	indexes map[uint64]*VaultTagIndexEntry
}

type VaultTagIndexEntry struct {
	count        int64
	index        *VaultMainIndex
	check_delete bool
}

type VaultTagListData struct {
	NextId uint64            `json:"next_id"`
	Tags   map[uint64]string `json:"tags"`
}

func (data *VaultTagListData) FindTag(tag_name string) (bool, uint64) {
	parsedName := strings.ToLower(ParseTagName(tag_name))
	for key, val := range data.Tags {
		if strings.ToLower(val) == parsedName {
			return true, key
		}
	}
	return false, 0
}

func ParseTagName(name string) string {
	return strings.Trim(
		strings.ReplaceAll(
			strings.ReplaceAll(name, "\n", " "),
			"\r",
			"",
		),
		" ",
	)
}

func ResolveTagIndexFilePath(base_path string, tag_id uint64) string {
	return path.Join(base_path, "tags", "tag_"+fmt.Sprint(tag_id)+".index")
}

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

func (tm *VaultTagManager) readData(key []byte) (*VaultTagListData, error) {
	if _, err := os.Stat(tm.tag_list_file); err == nil {
		// Load file
		b, err := ioutil.ReadFile(tm.tag_list_file)

		if err != nil {
			return nil, err
		}

		bJSON, err := decryptFileContents(b, key)

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

func (tm *VaultTagManager) ReadList(key []byte) (*VaultTagListData, error) {
	tm.tag_list_lock.StartRead() // Request read
	defer tm.tag_list_lock.EndRead()

	return tm.readData(key)
}

func (tm *VaultTagManager) StartWrite(key []byte) (*VaultTagListData, error) {
	tm.tag_list_lock.RequestWrite() // Request write

	return tm.readData(key)
}

func (tm *VaultTagManager) EndWrite(data *VaultTagListData, key []byte) error {
	defer tm.tag_list_lock.EndWrite()

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
	tm.tag_list_lock.StartWrite()

	err = os.Rename(tmpFile, tm.tag_list_file)

	return err
}

func (tm *VaultTagManager) CancelWrite() {
	tm.tag_list_lock.EndWrite()
}

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

func (tm *VaultTagManager) AddTagToList(tag_name string, key []byte) (uint64, error) {
	parsedTagName := ParseTagName(tag_name)

	data, err := tm.StartWrite(key)

	if err != nil {
		return 0, err
	}

	found, tag_id := data.FindTag(parsedTagName)

	if found {
		return tag_id, nil // Already exists
	}

	newIdForTag := data.NextId
	data.NextId++

	data.Tags[newIdForTag] = parsedTagName

	err = tm.EndWrite(data, key)

	return newIdForTag, err
}

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

func (tm *VaultTagManager) TagMedia(media_id uint64, tag_name string, key []byte) error {
	tag_id, err := tm.AddTagToList(tag_name, key)

	if err != nil {
		return err
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	r, err := indexFile.StartWrite()

	if err != nil {
		return err
	}

	added, _, err := r.file.AddValue(media_id)

	if err != nil {
		indexFile.CancelWrite(r)
		return err
	}

	if !added {
		// Already tagged
		indexFile.CancelWrite(r)
		return nil
	}

	err = indexFile.EndWrite(r)

	return err
}

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

func (tm *VaultTagManager) ListTaggedMedia(tag_name string, key []byte, skip int64, limit int64, reverse bool) ([]uint64, error) {
	tagList, err := tm.ReadList(key)

	if err != nil {
		return nil, err
	}

	found, tag_id := tagList.FindTag(tag_name)

	if !found {
		return make([]uint64, 0), nil
	}

	indexFile, err := tm.AcquireIndexFile(tag_id)

	if err != nil {
		return nil, err
	}

	defer tm.ReleaseIndexFile(tag_id, false, key)

	f, err := indexFile.StartRead()

	if err != nil {
		return nil, err
	}

	var values []uint64

	if reverse {
		values, err = f.ListValuesReverse(skip, limit)

		if err != nil {
			indexFile.EndRead(f)
			return nil, err
		}
	} else {
		values, err = f.ListValues(skip, limit)

		if err != nil {
			indexFile.EndRead(f)
			return nil, err
		}
	}

	indexFile.EndRead(f)

	return values, nil
}
