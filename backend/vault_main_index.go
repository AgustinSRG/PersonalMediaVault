// Vault main index manager

package main

import (
	"errors"
	"os"
)

// Manages an index file of the vault
type VaultMainIndex struct {
	file string         // path to the index file
	lock *ReadWriteLock // Lock to control access of threads
}

// Resource given when a write operation is requested
type VaultIndexWriteResource struct {
	file *IndexedListFile // Reference to the opened file
	path string           // Path to a temp file where changes are pre-made
}

// Initializes the index file manager
// file - Path to the index file to manage
func (vmi *VaultMainIndex) Initialize(file string) error {
	vmi.file = file
	vmi.lock = CreateReadWriteLock()

	if _, err := os.Stat(file); err == nil {
		return nil // File exists
	} else if errors.Is(err, os.ErrNotExist) {
		// Create empty index

		f, err := OpenIndexedListForWriting(file)

		if err != nil {
			return err
		}

		defer f.Close()

		err = f.Initialize()

		if err != nil {
			return err
		}

		return nil
	} else {
		return err
	}
}

// Starts write operation
// Returns a temp resource to make the changes
func (vmi *VaultMainIndex) StartWrite() (*VaultIndexWriteResource, error) {
	vmi.lock.RequestWrite() // Request write

	// Make temp file
	tmpFile := GetTemporalFileName("index", true)

	// Copy file
	_, err := CopyFile(vmi.file, tmpFile)

	if err != nil {
		vmi.lock.EndWrite()
		return nil, err
	}

	// Open temp file for writing
	fd, err := OpenIndexedListForWriting(tmpFile)

	if err != nil {
		vmi.lock.EndWrite()
		return nil, err
	}

	result := VaultIndexWriteResource{
		file: fd,
		path: tmpFile,
	}

	return &result, nil
}

// Applies a write operation
// res - Temp file resource
func (vmi *VaultMainIndex) EndWrite(res *VaultIndexWriteResource) error {
	res.file.Close()

	vmi.lock.StartWrite()

	// Move temp file to original path
	err := RenameAndReplace(res.path, vmi.file)

	vmi.lock.EndWrite()

	return err
}

// Deletes index file
func (vmi *VaultMainIndex) Delete() error {
	vmi.lock.RequestWrite()
	vmi.lock.StartWrite()

	// Remove index
	err := os.Remove(vmi.file)

	vmi.lock.EndWrite()

	return err
}

// Cancels write operation
// res - Temp file resource
func (vmi *VaultMainIndex) CancelWrite(res *VaultIndexWriteResource) {
	res.file.Close()
	os.Remove(res.path)
	vmi.lock.EndWrite()
}

// Starts read operation
// Opens the file for reading and returns the reference
func (vmi *VaultMainIndex) StartRead() (*IndexedListFile, error) {
	vmi.lock.StartRead() // Request read

	// Open temp file for reading
	fd, err := OpenIndexedListForReading(vmi.file)

	if err != nil {
		vmi.lock.EndRead()
		return nil, err
	}

	return fd, nil
}

// Ends read operation
// fd - Opened file (it closes it)
func (vmi *VaultMainIndex) EndRead(fd *IndexedListFile) {
	fd.Close()
	vmi.lock.EndRead()
}

// Removes element from main index
// media_id - ID of the media file
func (vmi *VaultMainIndex) RemoveElement(media_id uint64) error {
	r, err := vmi.StartWrite()

	if err != nil {
		return err
	}

	removed, _, err := r.file.RemoveValue(media_id)

	if err != nil {
		vmi.CancelWrite(r)
		return err
	}

	if !removed {
		// Was not tagged, no change
		vmi.CancelWrite(r)
		return nil
	}

	err = vmi.EndWrite(r)

	return err
}
