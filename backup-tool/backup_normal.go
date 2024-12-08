// Regular backup

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Runs regular backup
func RunNormalBackup(vaultPath string, backupPath string, tmpFile string) {
	// Welcome
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ModeRegularBackup",
			Other: "Mode: Regular backup",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	// Find files to check

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "FindingFiles",
			Other: "Finding files...",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	totalEntries := make([]BackupEntry, 0)

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "main.index"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "credentials.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "media_ids.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tasks.json"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "albums.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tag_list.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "user_config.pmv"))

	tagFiles := findBackupEntries(vaultPath, backupPath, "./tags")
	totalEntries = append(totalEntries, tagFiles...)

	mediaFiles := findBackupEntries(vaultPath, backupPath, "./media")
	totalEntries = append(totalEntries, mediaFiles...)

	albumThumbnails := findBackupEntries(vaultPath, backupPath, "./thumb_album")
	totalEntries = append(totalEntries, albumThumbnails...)

	resetLineOverWrite()

	// Initialize backup (check files to be copied)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InitializeNotice",
			Other: "Initializing backup...",
		},
	})
	fmt.Fprint(os.Stderr, msg)

	work, err := initializeBackupWork(totalEntries)

	fmt.Fprint(os.Stderr, "\n")

	if err != nil {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
		return
	}

	// Copy files

	bytesCopied := int64(0)
	bytesToCopy := work.totalSize

	realBytesCopied := int64(0)

	progressInt := int64(0)
	prevProgress := int64(0)

	if bytesToCopy < 1 {
		bytesToCopy = 1
	}

	statFilesCopied := 0

	resetLineOverWrite()

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "MakeNotice",
			Other: "Making backup...",
		},
	})
	fmt.Fprint(os.Stderr, msg)

	for i := 0; i < len(work.entries); i++ {

		copied, err := backupFile(work.entries[i], tmpFile, progressInt, i+1, len(work.entries))

		if err != nil {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "BackupFailed",
					Other: "Backup failed!",
				},
			})
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
			return
		}

		bytesCopied += work.entries[i].size

		if copied {
			statFilesCopied++
			realBytesCopied += work.entries[i].size
		}

		progressInt = bytesCopied * 100 / bytesToCopy
		if prevProgress != progressInt {
			prevProgress = progressInt
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "BackupProgress",
					Other: "Making backup... ({{.Percent}}%) ({{.Current}} / {{.Total}})",
				},
				TemplateData: map[string]interface{}{
					"Percent": fmt.Sprint(progressInt),
					"Current": fmt.Sprint(i + 1),
					"Total":   fmt.Sprint(len(work.entries)),
				},
			})
			printLineOverWrite(msg)
		}
	}

	// Done

	fmt.Fprint(os.Stderr, "\n")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "BackupDone",
			Other: "Backup done. Total files copied: {{.FileCount}} ({{.Size}})",
		},
		TemplateData: map[string]interface{}{
			"FileCount": fmt.Sprint(statFilesCopied),
			"Size":      formatBytes(realBytesCopied),
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}

func backupFile(entry BackupEntryExtended, tmpFile string, generalProgress int64, current int, total int) (copied bool, err error) {
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
			fmt.Fprintln(os.Stderr, "\n"+msg)
			return false, err
		}
	}

	// Make sure folder exists
	err = os.MkdirAll(entry.backupPath, FOLDER_PERMISSION)

	if err != nil {
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
		return false, err
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
			return false, err
		}
	}

	if fileInfoBackup == nil || fileInfo.ModTime().UnixMilli() > fileInfoBackup.ModTime().UnixMilli() {
		// Backup file does not exists, or it's older, copy it
		_, err = CopyFile(entry.original, entry.backupFile, tmpFile, generalProgress, current, total)

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
			fmt.Fprintln(os.Stderr, "\n"+msg)
			return false, err
		}

		return true, nil
	} else {
		return false, nil
	}
}

// Copy file
func CopyFile(src string, dst string, tmpFile string, generalProgress int64, current int, total int) (int64, error) {
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

	destination, err := os.Create(tmpFile)
	if err != nil {
		return 0, err
	}

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

	if err != nil {
		destination.Close()
		os.Remove(tmpFile)
		return 0, err
	}

	err = destination.Close()

	if err != nil {
		os.Remove(tmpFile)
		return 0, err
	}

	err = RenameAndReplace(tmpFile, dst)

	if err != nil {
		os.Remove(tmpFile)
		return 0, err
	}

	return written, nil
}
