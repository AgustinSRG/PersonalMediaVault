// Utils

package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

const (
	FILE_PERMISSION   = 0600 // Read/Write
	FOLDER_PERMISSION = 0700 // Read/Write/Run
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
