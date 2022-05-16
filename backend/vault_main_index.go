// Vault main index manager

package main

import (
	"errors"
	"os"
)

type VaultMainIndex struct {
	file string
	lock *ReadWriteLock
}

type VaultIndexWriteResource struct {
	file *IndexedListFile
	path string
}

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
func (vmi *VaultMainIndex) StartWrite() (*VaultIndexWriteResource, error) {
	vmi.lock.RequestWrite() // Request write

	// Make temp file
	tmpFile := GetTemporalFileName("index")

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
func (vmi *VaultMainIndex) EndWrite(res *VaultIndexWriteResource) error {
	res.file.Close()

	vmi.lock.StartWrite()

	// Move temp file to original apth
	err := os.Rename(res.path, vmi.file)

	vmi.lock.EndWrite()

	return err
}

// Deletes index
func (vmi *VaultMainIndex) Delete() error {
	vmi.lock.RequestWrite()
	vmi.lock.StartWrite()

	// Remove index
	err := os.Remove(vmi.file)

	vmi.lock.EndWrite()

	return err
}

// Cancels write operation
func (vmi *VaultMainIndex) CancelWrite(res *VaultIndexWriteResource) {
	res.file.Close()
	os.Remove(res.path)
	vmi.lock.EndWrite()
}

// Starts read operation
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
func (vmi *VaultMainIndex) EndRead(fd *IndexedListFile) {
	fd.Close()
	vmi.lock.EndRead()
}
