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
	temp_files_path    = "./temp"
	temp_files_prefix  = "pmv_tmp_"
	temp_files_lock    = &sync.Mutex{}
	temp_files_counter = 0
)

// Initialize
func SetTempFilesPath(path string) {
	temp_files_path = path

	// Create path if not exists
	os.MkdirAll(temp_files_path, FOLDER_PERMISSION)

	// Create unique prefix for each execution
	timeNow := time.Now().UTC().UnixMilli()
	temp_files_prefix = "pmv_tmp_" + fmt.Sprint(timeNow) + "_"
}

// Clears temp path on application exit
func ClearTemporalFilesPath() {
	os.RemoveAll(temp_files_path)
	os.MkdirAll(temp_files_path, FOLDER_PERMISSION)
}

// Gets a name for a temporal file
func GetTemporalFileName(extension string) string {
	temp_files_lock.Lock()

	temp_files_counter++
	fileName := temp_files_prefix + fmt.Sprint(temp_files_counter)

	temp_files_lock.Unlock()

	if extension != "" {
		fileName += "." + extension
	}

	return path.Join(temp_files_path, fileName)
}

// Creates a temporal folder and returns the path
func GetTemporalFolder() (string, error) {
	temp_files_lock.Lock()

	temp_files_counter++
	folderName := temp_files_prefix + fmt.Sprint(temp_files_counter)

	temp_files_lock.Unlock()

	folderPath := path.Join(temp_files_path, folderName)

	err := os.MkdirAll(folderPath, FOLDER_PERMISSION)

	if err != nil {
		return "", err
	}

	return folderPath, nil
}

// Wipes file to prevent recovery (secure delete)
func WipeTemporalFile(file string) {
	f, err := os.OpenFile(file, os.O_RDWR, FILE_PERMISSION)

	if err != nil {
		LogError(err)
		os.Remove(file)
		return
	}

	defer f.Close()
	defer os.Remove(file)

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

	// Number of chunks

	chunkCount := fileSize / int64(len(fileChunk))

	if fileSize%int64(len(fileChunk)) != 0 {
		chunkCount++
	}

	// Overwrite file

	for i := int64(0); i < chunkCount; i++ {
		f.Write(fileChunk)
	}
}
