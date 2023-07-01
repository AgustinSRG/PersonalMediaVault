// Tool to pack multiple files into the same file

// File structure
//  - Header (8 bytes):
//      - Number of files (Long unsigned big endian) (8 bytes)
//  - Files table (Table of rows of 16 bytes, one per file)
//      - Start position of the file (Long unsigned big endian) (8 bytes)
//      - File length (Long unsigned big endian) (8 bytes)
//  - Body: Files data, consistent with the files table

package main

import (
	"encoding/binary"
	"errors"
	"os"
)

//////////////////////////
//     WRITE STREAM    //
/////////////////////////

// Write stream data
type MultiFilePackWriteStream struct {
	f                   *os.File // File descriptor
	file_count          int64    // Number of files contained in the file
	current_write_index int64    // Index of the current file being written
	current_write_pt    int64    // Position of the cursor to write the next file
}

// Creates stream to write multiple files in a packed file
// file - Path of the file
func CreateMultiFilePackWriteStream(file string) (*MultiFilePackWriteStream, error) {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := MultiFilePackWriteStream{
		f:                   f,
		file_count:          0,
		current_write_index: 0,
	}

	return &i, nil
}

// Initializes write stream (must be called before writting any files)
// file_count - Number of files to write
func (file *MultiFilePackWriteStream) Initialize(file_count int64) error {
	file.file_count = file_count

	// Set the size of the file
	err := file.f.Truncate(8 + (16)*file_count)
	if err != nil {
		return err
	}

	// Rewind to the start of the file
	_, err = file.f.Seek(0, 0)

	if err != nil {
		return err
	}

	// Write size
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(file_count))

	_, err = file.f.Write(b)
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint64(b, 0)

	// Write default values for each file

	for i := int64(0); i < file_count; i++ {
		_, err = file.f.Write(b)
		if err != nil {
			return err
		}

		_, err = file.f.Write(b)
		if err != nil {
			return err
		}
	}

	file.current_write_index = 0
	file.current_write_pt = 8 + (16)*file_count

	return nil
}

// Writes a file into the packed file
// content - Content of the file
// Calling this increases the current file index
func (file *MultiFilePackWriteStream) PutFile(content []byte) error {
	// Save metadata
	_, err := file.f.Seek(8+file.current_write_index*16, 0)

	if err != nil {
		return err
	}

	b := make([]byte, 8)

	// Write start pointer
	binary.BigEndian.PutUint64(b, uint64(file.current_write_pt))
	_, err = file.f.Write(b)
	if err != nil {
		return err
	}

	// Write length
	binary.BigEndian.PutUint64(b, uint64(len(content)))
	_, err = file.f.Write(b)
	if err != nil {
		return err
	}

	// Write data

	_, err = file.f.Seek(file.current_write_pt, 0)

	if err != nil {
		return err
	}

	_, err = file.f.Write(content)
	if err != nil {
		return err
	}

	file.current_write_index++
	file.current_write_pt += int64(len(content))

	return nil
}

// Closes the stream
func (file *MultiFilePackWriteStream) Close() error {
	return file.f.Close()
}

//////////////////////////
//     READ STREAM     //
/////////////////////////

// Read stream to retreive files from a packed file
type MultiFilePackReadStream struct {
	f          *os.File // File descriptor
	file_count int64    // Number of files inside the packed file
}

// Creates read stream to get files from a packed file
// file - path to the file
func CreateMultiFilePackReadStream(file string) (*MultiFilePackReadStream, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return nil, err
	}

	i := MultiFilePackReadStream{
		f: f,
	}

	b := make([]byte, 8)

	_, err = f.Read(b)

	if err != nil {
		f.Close()
		return nil, err
	}

	i.file_count = int64(binary.BigEndian.Uint64(b))

	return &i, nil
}

// Gets a file
// index - file index
// Returns the file data
func (file *MultiFilePackReadStream) GetFile(index int64) ([]byte, error) {
	if index < 0 || index >= file.file_count {
		return nil, errors.New("Index out of bounds")
	}

	// Fetch metadata of the file
	_, err := file.f.Seek(8+index*16, 0)

	if err != nil {
		return nil, err
	}

	ptBytes := make([]byte, 8)
	lenBytes := make([]byte, 8)

	_, err = file.f.Read(ptBytes)

	if err != nil {
		return nil, err
	}

	_, err = file.f.Read(lenBytes)

	if err != nil {
		return nil, err
	}

	pt := int64(binary.BigEndian.Uint64(ptBytes))

	_, err = file.f.Seek(pt, 0)

	if err != nil {
		return nil, err
	}

	l := int64(binary.BigEndian.Uint64(lenBytes))

	data := make([]byte, l)

	_, err = file.f.Read(data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

// Closes the read stream
func (file *MultiFilePackReadStream) Close() {
	file.f.Close()
}
