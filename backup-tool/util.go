// Utilities

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	FILE_PERMISSION   = 0600 // Read/Write
	FOLDER_PERMISSION = 0700 // Read/Write/Run
)

// Copy file
func CopyFile(src string, dst string, generalProgress int64, current int, total int) (int64, error) {
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

	buf := make([]byte, 32*1000)

	written := int64(0)
	progressInt := int64(0)
	prevProgress := int64(0)
	size := sourceFileStat.Size()

	if size < 1 {
		size = 1
	}

	for {
		nr, er := source.Read(buf)
		if nr > 0 {
			nw, ew := destination.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = errors.New("short write")
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}

		progressInt = int64(written) * 100 / size
		if prevProgress != progressInt {
			prevProgress = progressInt
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "BackupCopyProgress",
					Other: "Making backup... ({{.Percent}}%) ({{.Current}} / {{.Total}}) - {{.File}} ({{.PercentFile}}%)",
				},
				TemplateData: map[string]interface{}{
					"Percent":     fmt.Sprint(generalProgress),
					"Current":     fmt.Sprint(current),
					"Total":       fmt.Sprint(total),
					"File":        dst,
					"PercentFile": fmt.Sprint(progressInt),
				},
			})
			printLineOverWrite(msg)
		}
	}

	return written, err
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

func backupFile(entry BackupEntryExtended, generalProgress int64, current int, total int) (copied bool, err error) {
	fileInfo, err := os.Stat(entry.original)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil // File was deleted / never existed
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
			fmt.Println("\n" + msg)
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
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Error",
					Other: "Error: {{.Message}}",
				},
				TemplateData: map[string]interface{}{
					"Message": err.Error(),
				},
			})
			fmt.Println("\n" + msg)
			return false, err
		}
	}

	if fileInfoBackup == nil || fileInfo.ModTime().UnixMilli() > fileInfoBackup.ModTime().UnixMilli() || fileInfo.Size() != fileInfoBackup.Size() {
		// Backup file does not exists, or it's older, copy it
		_, err = CopyFile(entry.original, entry.backupFile, generalProgress, current, total)

		if err != nil {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorCopyFile",
					Other: "Error copying file {{.File}} | Error: {{.Message}}",
				},
				TemplateData: map[string]interface{}{
					"File":    entry.original,
					"Message": err.Error(),
				},
			})
			fmt.Println("\n" + msg)
			return false, err
		}

		return true, nil
	} else {
		return false, nil
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

var PreviousLine = ""

func printLineOverWrite(line string) {
	paddedLine := line
	for len(PreviousLine) > len(paddedLine) {
		paddedLine = paddedLine + " "
	}

	PreviousLine = line

	fmt.Print("\r" + paddedLine)
}

func resetLineOverWrite() {
	PreviousLine = ""
}
