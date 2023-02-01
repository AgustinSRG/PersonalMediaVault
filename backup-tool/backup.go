// Backup actions

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

// Backup entry
type BackupEntry struct {
	original   string
	backupPath string
	backupFile string
}

func makeBackupEntry(vaultPath string, backupPath string, relativePath string, file string) BackupEntry {
	return BackupEntry{
		original:   path.Join(vaultPath, relativePath, file),
		backupPath: path.Join(backupPath, relativePath),
		backupFile: path.Join(backupPath, relativePath, file),
	}
}

func findBackupEntries(vaultPath string, backupPath string, relativePath string) []BackupEntry {
	pathToRead := path.Join(vaultPath, relativePath)

	readInfo, err := os.ReadDir(pathToRead)

	if err != nil {
		fmt.Println("\nError reading path " + pathToRead + " | Error: " + err.Error())
		return make([]BackupEntry, 0)
	}

	result := make([]BackupEntry, 0)

	for i := 0; i < len(readInfo); i++ {
		if readInfo[i].Type().IsDir() {
			entries := findBackupEntries(vaultPath, backupPath, path.Join(relativePath, readInfo[i].Name()))

			result = append(result, entries...)
		} else if readInfo[i].Type().IsRegular() {
			// Regular file
			entry := makeBackupEntry(vaultPath, backupPath, relativePath, readInfo[i].Name())

			result = append(result, entry)
		}
	}

	return result
}

func backupFile(entry BackupEntry) (copied bool, err error) {
	fileInfo, err := os.Stat(entry.original)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil // File was deleted / never existed
		} else {
			fmt.Println("\nError fetching info of " + entry.original + " | Error: " + err.Error())
			return false, err
		}
	}

	// Make sure folder exists
	os.MkdirAll(entry.backupPath, FOLDER_PERMISSION)

	fileInfoBackup, err := os.Stat(entry.backupFile)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fileInfoBackup = nil
		} else {
			fmt.Println("\nError: " + err.Error())
			return false, err
		}
	}

	if fileInfoBackup == nil || fileInfo.ModTime().UnixMilli() > fileInfoBackup.ModTime().UnixMilli() {
		// Backup file does not exists, or it's older, copy it
		_, err = CopyFile(entry.original, entry.backupFile)

		if err != nil {
			fmt.Println("\nError copying file " + entry.original + " | Error: " + err.Error())
			return false, err
		}

		return true, nil
	} else {
		return false, nil
	}
}
