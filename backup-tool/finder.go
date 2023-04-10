// Backup actions

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/nicksnyder/go-i18n/v2/i18n"
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
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrorReadPath",
				Other: "Error reading path {{.Path}} | Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Path":    pathToRead,
				"Message": err.Error(),
			},
		})
		fmt.Println("\n" + msg)
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
