// Utilities

package main

import (
	"errors"
	"io"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	FILE_PERMISSION   = 0600 // Read/Write
	FOLDER_PERMISSION = 0700 // Read/Write/Run
)

// Copy file
func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorNotRegularFile",
				Other: "{{.File}} is not a regular file",
			},
			TemplateData: map[string]interface{}{
				"File": src,
			},
		})
		return 0, errors.New(msg)
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

func CheckFileExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false
	} else {
		return false
	}
}
