// Home page configuration

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

// Media element
const HOME_PAGE_ELEMENT_TYPE_MEDIA = 0

// Album element
const HOME_PAGE_ELEMENT_TYPE_ALBUM = 1

// Home page element
type HomePageElement struct {
	// The element type (media, album)
	ElementType uint8 `json:"t,omitempty"`

	// The ID of the media or the album
	Id uint64 `json:"i"`
}

// Custom group
const HOME_PAGE_GROUP_CUSTOM = 0

// Special group: Recent uploaded media
const HOME_PAGE_GROUP_RECENT_MEDIA = 1

// Special group: Recent uploaded albums
const HOME_PAGE_GROUP_RECENT_ALBUMS = 2

// Checks if the group type is valid
func validateGroupType(groupType uint8) bool {
	return groupType <= HOME_PAGE_GROUP_RECENT_ALBUMS
}

// Group of elements for the home page
type HomePageGroup struct {
	// Group unique ID
	Id uint64 `json:"id"`

	// Type of group
	Type uint8 `json:"type,omitempty"`

	// Name of the group, in order to display it
	Name string `json:"name,omitempty"`

	// List of elements of the group
	Elements []HomePageElement `json:"elements,omitempty"`
}

// Configuration for the home page
type HomePageConfiguration struct {
	// List of groups
	Groups []HomePageGroup `json:"groups"`

	// ID for the next group
	NextId uint64 `json:"next_id"`
}

// Initializes configuration
func (config *HomePageConfiguration) Initialize() {
	config.Groups = []HomePageGroup{
		{
			Id:   0,
			Type: HOME_PAGE_GROUP_RECENT_MEDIA,
		},
		{
			Id:   1,
			Type: HOME_PAGE_GROUP_RECENT_ALBUMS,
		},
	}
	config.NextId = 2
}

// Finds group by ID
// Returns the index of the group, or -1 if not found
func (config *HomePageConfiguration) FindGroup(id uint64) int {
	for i, g := range config.Groups {
		if g.Id == id {
			return i
		}
	}

	return -1
}

// Manager for home page config
type HomePageConfigManager struct {
	file  string                 // Home page config file
	cache *HomePageConfiguration // Cache
	lock  *ReadWriteLock         // Lock to control access to the file
}

// Initializes home page config manager
// base_path - Vault path
func (uc *HomePageConfigManager) Initialize(base_path string) {
	uc.cache = nil
	uc.file = path.Join(base_path, "home_page.pmv")
	uc.lock = CreateReadWriteLock()
}

// Reads home page config
// key - Vault decryption key
// Returns home page config data
func (hpc *HomePageConfigManager) Read(key []byte) (*HomePageConfiguration, error) {
	hpc.lock.StartRead()
	defer hpc.lock.EndRead()

	return hpc.readInternal(key)
}

// Reads home page config
// key - Vault decryption key
// Returns home page config data
func (hpc *HomePageConfigManager) readInternal(key []byte) (*HomePageConfiguration, error) {
	if hpc.cache != nil {
		return hpc.cache, nil
	}

	if _, err := os.Stat(hpc.file); err == nil {
		// Load file
		b, err := os.ReadFile(hpc.file)

		if err != nil {
			return nil, err
		}

		bJSON, err := encrypted_storage.DecryptFileContents(b, key)

		if err != nil {
			return nil, err
		}

		var mp HomePageConfiguration

		err = json.Unmarshal(bJSON, &mp)

		if err != nil {
			return nil, err
		}

		hpc.cache = &mp

		return &mp, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// Default config

		mp := HomePageConfiguration{}
		mp.Initialize()

		hpc.cache = &mp

		return &mp, nil
	} else {
		return nil, err
	}
}

// Creates a new group in home page
// key - The vault encryption key
// name - Name for the group
// groupType - Type of group
// prepend - True to insert at the beginning
func (hpc *HomePageConfigManager) CreateGroup(key []byte, name string, groupType uint8, prepend bool) (uint64, error) {
	hpc.lock.RequestWrite()
	defer hpc.lock.EndWrite()

	config, err := hpc.readInternal(key)

	if err != nil {
		return 0, err
	}

	groupId := config.NextId
	config.NextId++

	if config.Groups == nil {
		config.Groups = make([]HomePageGroup, 0)
	}

	group := HomePageGroup{
		Id:       groupId,
		Type:     groupType,
		Name:     name,
		Elements: make([]HomePageElement, 0),
	}

	if prepend {
		config.Groups = append([]HomePageGroup{group}, config.Groups...)
	} else {
		config.Groups = append(config.Groups, group)
	}

	return groupId, hpc.finishWrite(config, key)
}

// Moves a group in home page
// key - The vault encryption key
// id - Id of the group
// position - Position for the group to be moved
func (hpc *HomePageConfigManager) MoveGroup(key []byte, id uint64, position int) error {
	hpc.lock.RequestWrite()
	defer hpc.lock.EndWrite()

	config, err := hpc.readInternal(key)

	if err != nil {
		return err
	}

	if len(config.Groups) == 0 {
		return nil
	}

	groupPos := config.FindGroup(id)

	if groupPos == -1 || groupPos == position {
		return nil
	}

	group := config.Groups[groupPos]

	if position < 0 {
		position = 0
	}

	oldList := config.Groups

	if position > len(oldList) {
		position = len(oldList)
	}

	newList := make([]HomePageGroup, 0)

	j := 0 // Position in the new list

	for i := 0; i < len(oldList); i++ {
		if j == position {
			newList = append(newList, group)
			j++
		}
		if oldList[i].Id != id {
			newList = append(newList, oldList[i])
			j++
		}
	}

	if position >= len(newList) {
		newList = append(newList, group)
	}

	config.Groups = newList

	return hpc.finishWrite(config, key)
}

// Renames a group in home page
// key - The vault encryption key
// id - Id of the group
// name - New name for the group
func (hpc *HomePageConfigManager) RenameGroup(key []byte, id uint64, name string) error {
	hpc.lock.RequestWrite()
	defer hpc.lock.EndWrite()

	config, err := hpc.readInternal(key)

	if err != nil {
		return err
	}

	if len(config.Groups) == 0 {
		return nil
	}

	groupPos := config.FindGroup(id)

	if groupPos == -1 {
		return nil
	}

	config.Groups[groupPos].Name = name

	return hpc.finishWrite(config, key)
}

// Sets the element list of a group in the home page
// key - The vault encryption key
// id - Id of the group
// elements - List of elements
func (hpc *HomePageConfigManager) SetGroupElementList(key []byte, id uint64, elements []HomePageElement) error {
	hpc.lock.RequestWrite()
	defer hpc.lock.EndWrite()

	config, err := hpc.readInternal(key)

	if err != nil {
		return err
	}

	if len(config.Groups) == 0 {
		return nil
	}

	groupPos := config.FindGroup(id)

	if groupPos == -1 {
		return nil
	}

	config.Groups[groupPos].Elements = elements

	return hpc.finishWrite(config, key)
}

// Removes a group from the home page
// key - The vault encryption key
// id - Id of the group
// name - New name for the group
func (hpc *HomePageConfigManager) DeleteGroup(key []byte, id uint64) error {
	hpc.lock.RequestWrite()
	defer hpc.lock.EndWrite()

	config, err := hpc.readInternal(key)

	if err != nil {
		return err
	}

	if len(config.Groups) == 0 {
		return nil // Nothing to move
	}

	newList := make([]HomePageGroup, 0)

	for _, g := range config.Groups {
		if g.Id != id {
			newList = append(newList, g)
		}
	}

	config.Groups = newList

	return hpc.finishWrite(config, key)
}

// Writes home page configuration
// data - Data to write
// key - Vault encryption key
func (hpc *HomePageConfigManager) finishWrite(data *HomePageConfiguration, key []byte) error {
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
	hpc.lock.StartWrite()

	err = RenameAndReplace(tmpFile, hpc.file)

	hpc.cache = data

	return err
}
