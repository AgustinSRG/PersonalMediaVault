// Temporal files manager

package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

var (
	temp_files_path             = "./temp"
	unencrypted_temp_files_path = "./temp"

	temp_files_prefix  = "pmv_tmp_"
	temp_files_lock    = &sync.Mutex{}
	temp_files_counter = 0
)

// Initialize
func SetTempFilesPath(tempPath string, unencryptedTempPath string) {
	temp_files_path = tempPath
	unencrypted_temp_files_path = unencryptedTempPath

	// Create path if not exists
	os.MkdirAll(temp_files_path, FOLDER_PERMISSION)
	os.MkdirAll(unencrypted_temp_files_path, FOLDER_PERMISSION)

	// Create unique prefix for each execution
	timeNow := time.Now().UTC().UnixMilli()
	temp_files_prefix = "pmv_tmp_" + fmt.Sprint(timeNow) + "_"
}

// Clears temp path on application exit
func ClearTemporalFilesPath() {
	os.MkdirAll(temp_files_path, FOLDER_PERMISSION)

	entries, err := os.ReadDir(temp_files_path)

	if err != nil {
		LogError(err)
		return
	}

	for i := 0; i < len(entries); i++ {
		if entries[i].Type().IsRegular() {
			WipeTemporalFile(path.Join(temp_files_path, entries[i].Name()))
		} else if entries[i].Type().IsDir() {
			WipeTemporalPath(path.Join(temp_files_path, entries[i].Name()))
		}
	}

	os.MkdirAll(unencrypted_temp_files_path, FOLDER_PERMISSION)

	entries, err = os.ReadDir(unencrypted_temp_files_path)

	if err != nil {
		LogError(err)
		return
	}

	for i := 0; i < len(entries); i++ {
		if entries[i].Type().IsRegular() {
			WipeTemporalFile(path.Join(unencrypted_temp_files_path, entries[i].Name()))
		} else if entries[i].Type().IsDir() {
			WipeTemporalPath(path.Join(unencrypted_temp_files_path, entries[i].Name()))
		}
	}
}

// Gets a name for a temporal file
func GetTemporalFileName(extension string, encrypted bool) string {
	temp_files_lock.Lock()

	temp_files_counter++
	fileName := temp_files_prefix + fmt.Sprint(temp_files_counter)

	temp_files_lock.Unlock()

	if extension != "" {
		fileName += "." + extension
	}

	var baseFolder string

	if encrypted {
		baseFolder = temp_files_path
	} else {
		baseFolder = unencrypted_temp_files_path
	}

	return path.Join(baseFolder, fileName)
}

// Creates a temporal folder and returns the path
func GetTemporalFolder(encrypted bool) (string, error) {
	temp_files_lock.Lock()

	temp_files_counter++
	folderName := temp_files_prefix + fmt.Sprint(temp_files_counter)

	temp_files_lock.Unlock()

	var baseFolder string

	if encrypted {
		baseFolder = temp_files_path
	} else {
		baseFolder = unencrypted_temp_files_path
	}

	folderPath := path.Join(baseFolder, folderName)

	err := os.MkdirAll(folderPath, FOLDER_PERMISSION)

	if err != nil {
		return "", err
	}

	return folderPath, nil
}

// Wipes file to prevent recovery (secure delete)
func WipeTemporalFile(file string) {
	f, err := os.OpenFile(file, os.O_WRONLY, FILE_PERMISSION)

	if err != nil {
		LogError(err)
		os.Remove(file)
		return
	}

	defer func() {
		f.Close()
		os.Remove(file)
	}()

	fileInfo, err := f.Stat()
	if err != nil {
		LogError(err)
		return
	}

	fileSize := fileInfo.Size()
	fileChunk := make([]byte, 1024*1024)

	// Fill chunk with 0
	for i := 0; i < len(fileChunk); i++ {
		fileChunk[i] = 0
	}

	// Overwrite

	bytesWritten := int64(0)

	for bytesWritten < fileSize {
		bytesToWrite := int64(len(fileChunk))

		if bytesToWrite > (fileSize - bytesWritten) {
			bytesToWrite = fileSize - bytesWritten
		}

		f.Write(fileChunk[:bytesToWrite])

		bytesWritten += bytesToWrite
	}
}

func WipeTemporalPath(p string) {
	entries, err := os.ReadDir(p)

	if err != nil {
		LogError(err)
		return
	}

	for i := 0; i < len(entries); i++ {
		if entries[i].Type().IsRegular() {
			WipeTemporalFile(path.Join(p, entries[i].Name()))
		}
	}

	os.Remove(p)
}
