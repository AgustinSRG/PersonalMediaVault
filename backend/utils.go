// Utils

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	swap_files "github.com/AgustinSRG/go-swap-files"
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
	defer source.Close() //nolint:errcheck

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close() //nolint:errcheck
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

// Swaps 2 files
// If it fails, tries again up to 3 times, waiting 500 ms (this is to wait for any other program to unlock the file)
// file1 - File 1
// file2 - File 2
// swapFile - Intermediate file
// returns the error
func SwapFiles(file1 string, file2 string, swapFile string) error {
	retriesLeft := 3
	var err error = nil

	for retriesLeft > 0 {
		err = swap_files.SwapFiles(file1, file2, swapFile)

		if err == nil {
			return nil
		}

		if _, err := os.Stat(file1); errors.Is(err, os.ErrNotExist) {
			if _, err := os.Stat(swapFile); err == nil {
				// Step 3 failed
				// Restore file1 if swapFile exists but file1 got removed
				err = os.Rename(swapFile, file1)

				if err == nil {
					return nil
				}
			}
		} else if _, err := os.Stat(file2); errors.Is(err, os.ErrNotExist) {
			if _, err := os.Stat(swapFile); err == nil {
				// Step 2 failed
				// Restore file2 if swapFile exists but file2 got removed
				_ = os.Rename(swapFile, file2)
			}
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

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}
