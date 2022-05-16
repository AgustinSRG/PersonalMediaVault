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
	count int64
	index *VaultMainIndex
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
		count: 1,
		index: &newIndex,
	}

	return &newIndex, nil
}

func (tm *VaultTagManager) ReleaseIndexFile(tag_id uint64) {
	tm.lock.Lock()
	defer tm.lock.Unlock()

	if tm.indexes[tag_id] != nil {
		tm.indexes[tag_id].count--

		if tm.indexes[tag_id].count <= 0 {
			delete(tm.indexes, tag_id)
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
