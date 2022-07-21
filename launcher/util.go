// Utils

package main

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	FILE_PERMISSION   = 0600 // Read/Write
	FOLDER_PERMISSION = 0700 // Read/Write/Run
)

func fileExists(file string) bool {
	sourceFileStat, err := os.Stat(file)
	if err != nil {
		return false
	}

	return sourceFileStat.Mode().IsRegular()
}

func folderExists(folder string) bool {
	sourceFileStat, err := os.Stat(folder)
	if err != nil {
		return false
	}

	return sourceFileStat.Mode().IsDir()
}

func getDirName() string {
	ex, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(ex)
}

func getBinaryFileName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	} else {
		return name
	}
}
