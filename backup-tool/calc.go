// Calculate work

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type BackupEntryExtended struct {
	original   string
	backupPath string
	backupFile string

	size int64
}

type BackupWork struct {
	entries   []BackupEntryExtended
	totalSize int64
}

func initializeBackupWork(foundEntries []BackupEntry) (*BackupWork, error) {
	work := BackupWork{
		entries:   make([]BackupEntryExtended, 0),
		totalSize: 0,
	}

	progressInt := int64(0)
	prevProgress := int64(0)

	for i := 0; i < len(foundEntries); i++ {

		entry := foundEntries[i]

		fileInfo, err := os.Stat(entry.original)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			} else {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorFetchFile",
						Other: "Error fetching info of {{.File}} | Error: {{.Message}}",
					},
					TemplateData: map[string]interface{}{
						"File":    entry.original,
						"Message": err.Error(),
					},
				})
				fmt.Fprintln(os.Stderr, "\n"+msg)
				return nil, err
			}
		}

		fileInfoBackup, err := os.Stat(entry.backupFile)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fileInfoBackup = nil
			} else {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Error",
						Other: "Error: {{.Message}}",
					},
					TemplateData: map[string]interface{}{
						"Message": err.Error(),
					},
				})
				fmt.Fprintln(os.Stderr, "\n"+msg)
				return nil, err
			}
		}

		if fileInfoBackup == nil || fileInfo.ModTime().UnixMilli() > fileInfoBackup.ModTime().UnixMilli() || fileInfo.Size() != fileInfoBackup.Size() {
			work.entries = append(work.entries, BackupEntryExtended{
				original:   entry.original,
				backupPath: entry.backupPath,
				backupFile: entry.backupFile,
				size:       fileInfo.Size(),
			})
			work.totalSize += fileInfo.Size()
		}

		progressInt = int64(i+1) * 100 / int64(len(foundEntries))
		if prevProgress != progressInt {
			prevProgress = progressInt
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "InitializeProgress",
					Other: "Initializing backup... ({{.Percent}}%) ({{.Current}} / {{.Total}})",
				},
				TemplateData: map[string]interface{}{
					"Percent": fmt.Sprint(progressInt),
					"Current": fmt.Sprint(i + 1),
					"Total":   fmt.Sprint(len(foundEntries)),
				},
			})
			printLineOverWrite(msg)
		}
	}

	return &work, nil
}
