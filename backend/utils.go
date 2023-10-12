// Utils

package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	FILE_PERMISSION          = 0600 // Read/Write
	FOLDER_PERMISSION        = 0700 // Read/Write/Run
	ENCRYPTED_BLOCK_MAX_SIZE = 5 * 1024 * 1024
)

// Copy file
// src - Source file
// dst - Destination path
// Returns the number of bytes copied
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

// Gets extension from file name
// fileName - File name
func GetExtensionFromFileName(fileName string) string {
	parts := strings.Split(fileName, ".")

	if len(parts) > 1 {
		ext := strings.ToLower(parts[len(parts)-1])

		r := regexp.MustCompile("[^a-z0-9]+")

		ext = r.ReplaceAllString(ext, "")

		if ext != "" {
			return ext
		} else {
			return "bin"
		}
	} else {
		return "bin"
	}
}

// Removes extension from file name
// fileName - File name
func GetNameFromFileName(fileName string) string {
	parts := strings.Split(fileName, ".")

	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	} else {
		return fileName
	}
}

// Renames and replaces file (Atomic)
// If it fails, tries again up to 3 times, waiting 500 ms (this is to wait for any other program to unlock the file)
// tmpFile - The temporal file to move
// destFile - The destination file name
// returns the error
func RenameAndReplace(tmpFile string, destFile string) error {
	retriesLeft := 3
	var err error = nil

	for retriesLeft > 0 {
		err = os.Rename(tmpFile, destFile)

		if err == nil {
			return nil
		}

		retriesLeft--

		time.Sleep(500 * time.Millisecond)
	}

	return err
}

// Checks if an album list has repeated elements, and removes them
// list - The media IDs list
// Returns the list without repeated elements
func AlbumListPruneRepeatedElements(list []uint64) []uint64 {
	m := make(map[uint64]struct{})
	res := make([]uint64, 0)

	for i := 0; i < len(list); i++ {
		e := list[i]
		_, repeated := m[e]

		if repeated {
			continue
		}

		res = append(res, e)
		m[e] = struct{}{}
	}

	return res
}
