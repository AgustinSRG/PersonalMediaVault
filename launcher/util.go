// Utils

package main

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

const (
	FILE_PERMISSION    = 0600 // Read/Write
	FOLDER_PERMISSION  = 0700 // Read/Write/Run
	DEFAULT_CACHE_SIZE = 1024 // Default cache size
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

// Copy file
// src - Source file
// dst - Destination path
// Returns the number of bytes copied
func CopyFile(src, dst string) (int64, error) {
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

var SSE_MODEL_FILES []string = []string{
	"model_config.json",
	"open_clip_config.json",
	"tokenizer.json",
	"tokenizer_config.json",
	"special_tokens_map.json",
	"text.onnx",
	"text.onnx.data",
	"visual.onnx",
	"visual.onnx.data",
}

func ValidateSemanticSearchModel(p string) (bool, string) {
	for _, file := range SSE_MODEL_FILES {
		finalPath := path.Join(p, file)

		if !fileExists(finalPath) {
			return false, file
		}
	}

	return true, ""
}
