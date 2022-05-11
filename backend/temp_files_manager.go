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
	os.MkdirAll(temp_files_path, 0755)

	// Create unique prefix for each execution
	timeNow := time.Now().UTC().UnixMilli()
	temp_files_prefix = "pmv_tmp_" + fmt.Sprint(timeNow) + "_"
}

// Clears temp path on application exit
func ClearTemporalFilesPath() {
	os.RemoveAll(temp_files_path)
	os.MkdirAll(temp_files_path, 0755)
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

	err := os.MkdirAll(folderPath, 0755)

	if err != nil {
		return "", err
	}

	return folderPath, nil
}
