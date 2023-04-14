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
	temp_files_path             = "./temp" // Path for vault temp files
	unencrypted_temp_files_path = "./temp" // Path for unencrypted temp files (upload, ffmpeg, etc)

	temp_files_prefix  = "pmv_tmp_"
	temp_files_lock    = &sync.Mutex{}
	temp_files_counter = 0

	secure_temp_file_delete = os.Getenv("TEMP_FILE_DELETE_MODE") == "SECURE"
)

// Set vault temp files path
// tempPath - Temp files path
func SetTempFilesPath(tempPath string) {
	temp_files_path = tempPath

	// Create path if not exists
	err := os.MkdirAll(temp_files_path, FOLDER_PERMISSION)
	if err != nil {
		LogError(err)
	}

	// Create unique prefix for each execution
	timeNow := time.Now().UTC().UnixMilli()
	temp_files_prefix = "pmv_tmp_" + fmt.Sprint(timeNow) + "_"
}

// Set unencrypted temp files path
// unencryptedTempPath - Unencrypted temp files path
func SetUnencryptedTempFilesPath(unencryptedTempPath string) {
	unencrypted_temp_files_path = unencryptedTempPath

	// Create path if not exists
	err := os.MkdirAll(unencrypted_temp_files_path, FOLDER_PERMISSION)

	if err != nil {
		LogError(err)
	}
}

// Clears vault temp path
func ClearTemporalFilesPath() {
	err := os.RemoveAll(temp_files_path)

	if err != nil {
		LogError(err)
	}

	err = os.MkdirAll(temp_files_path, FOLDER_PERMISSION)

	if err != nil {
		LogError(err)
	}
}

// Clears all unencrypted temp files
func ClearUnencryptedTempFilesPath() {
	err := os.MkdirAll(unencrypted_temp_files_path, FOLDER_PERMISSION)

	if err != nil {
		LogError(err)
		return
	}

	entries, err := os.ReadDir(unencrypted_temp_files_path)

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
// extension - File extension (without the dot)
// encrypted - True to use the vault temp path, false to use the unencrypted temp path
// Returns the path to the file
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
// encrypted - True to use the vault temp path, false to use the unencrypted temp path
// Returns the path to the folder (creates it)
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
// file - File path
func WipeTemporalFile(file string) {
	if !secure_temp_file_delete {
		os.Remove(file)
		return
	}

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

		_, err = f.Write(fileChunk[:bytesToWrite])

		if err != nil {
			LogError(err)
			return
		}

		bytesWritten += bytesToWrite
	}
}

// Wipes unencrypted temp path (secure delete)
// p - Path
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
