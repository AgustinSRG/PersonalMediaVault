// Re-encrypted backup

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/term"
)

// Run re-encrypted backup
func RunReEncryptedBackup(vaultPath string, backupPath string, tmpFile string) {
	// Welcome
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ModeReEncryptedBackup",
			Other: "Mode: Re-encrypted backup",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	// Load credentials

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LoadingCredentials",
			Other: "Loading credentials...",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	credentialsEntry := makeBackupEntry(vaultPath, backupPath, "./", "credentials.json")

	sourceCredentials, err := ReadVaultCredentials(credentialsEntry.original)

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

	// Ask for root password

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "EnterRootPassword",
			Other: "Enter root vault account ({{.Account}}) password",
		},
		TemplateData: map[string]interface{}{
			"Account": sourceCredentials.User,
		},
	})
	fmt.Fprint(os.Stderr, msg+": ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
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
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "\n")

	password := strings.TrimSpace(string(bytePassword))

	// Check password

	if !CheckPassword(password, sourceCredentials.Method, sourceCredentials.PasswordHash, sourceCredentials.Salt) {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "WrongPasswordError",
				Other: "Error: Wrong password",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Get vault key

	sourceVaultKey, err := DecryptKey(password, sourceCredentials.Method, sourceCredentials.PasswordHash, sourceCredentials.Salt, sourceCredentials.EncryptedKey)

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
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Check the backup credentials

	var destVaultKey []byte

	if CheckFileExists(credentialsEntry.backupFile) {
		// Credentials exits for backup, check for the key (same password)
		destCredentials, err := ReadVaultCredentials(credentialsEntry.backupFile)

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

		if !CheckPassword(password, destCredentials.Method, destCredentials.PasswordHash, destCredentials.Salt) {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "WrongPasswordDestError",
					Other: "Error: Destination credentials file exists, but password is not the same",
				},
			})
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}

		destVaultKey, err = DecryptKey(password, destCredentials.Method, destCredentials.PasswordHash, destCredentials.Salt, destCredentials.EncryptedKey)

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
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}
	} else {
		// Create new credentials file

		newCredentials, err := MakeCredentials(sourceCredentials.User, password, sourceCredentials.VaultFingerprint)

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
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}

		destVaultKey, err = DecryptKey(password, newCredentials.Method, newCredentials.PasswordHash, newCredentials.Salt, newCredentials.EncryptedKey)

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
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}

		// Save credentials file

		err = newCredentials.WriteToFile(credentialsEntry.backupFile, tmpFile)

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
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}
	}

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

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "media_ids.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tasks.json"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "albums.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tag_list.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "user_config.pmv"))

	tagFiles := findBackupEntries(vaultPath, backupPath, "./tags")
	totalEntries = append(totalEntries, tagFiles...)

	mediaFiles := findBackupEntries(vaultPath, backupPath, "./media")
	totalEntries = append(totalEntries, mediaFiles...)

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

		copied, err := backupFileReEncrypt(work.entries[i], tmpFile, progressInt, i+1, len(work.entries), sourceVaultKey, destVaultKey)

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
			ID:    "BackupDoneReEncrypt",
			Other: "Backup done. Total files re-encrypted: {{.FileCount}} ({{.Size}})",
		},
		TemplateData: map[string]interface{}{
			"FileCount": fmt.Sprint(statFilesCopied),
			"Size":      formatBytes(realBytesCopied),
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}

func backupFileReEncrypt(entry BackupEntryExtended, tmpFile string, generalProgress int64, current int, total int, srcKey []byte, dstKey []byte) (copied bool, err error) {
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
		// Backup file does not exists, or it's older, copy it (with re-encryption)
		_, err = CopyFileReEncrypt(entry.original, entry.backupFile, tmpFile, generalProgress, current, total, srcKey, dstKey)

		if err != nil {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorReEncryptFile",
					Other: "Error re-encrypting file {{.File}} | Error: {{.Message}}",
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

func CopyFileReEncrypt(src string, dst string, tmpFile string, generalProgress int64, current int, total int, srcKey []byte, dstKey []byte) (int64, error) {
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

	fileName := filepath.Base(src)
	size := sourceFileStat.Size()

	if strings.HasSuffix(fileName, ".pmv") {
		// Fast re-encrypt
		err = FastReEncrypt(src, dst, tmpFile, srcKey, dstKey)

		if err != nil {
			return 0, err
		}

		return size, nil
	} else if strings.HasSuffix(fileName, ".pma") {
		// Asset
		if strings.HasPrefix(fileName, "m_") {
			// Multi-file
			return ReEncryptMultiAssetFile(src, dst, tmpFile, generalProgress, current, total, srcKey, dstKey)
		} else {
			// Single file
			return ReEncryptSingleAssetFile(src, dst, tmpFile, generalProgress, current, total, srcKey, dstKey)
		}
	} else {
		// Normal copy
		return CopyFile(src, dst, tmpFile, generalProgress, current, total)
	}
}

func FastReEncrypt(src string, dst string, tmpFile string, srcKey []byte, dstKey []byte) error {
	// Read source
	b, err := os.ReadFile(src)

	if err != nil {
		return err
	}

	// Decrypt
	decrypted, err := decryptFileContents(b, srcKey)

	if err != nil {
		return err
	}

	// Re-Encrypt

	encrypted, err := encryptFileContents(decrypted, AES256_ZIP, dstKey)

	if err != nil {
		return err
	}

	// Save to temp file

	err = os.WriteFile(tmpFile, encrypted, FILE_PERMISSION)
	if err != nil {
		return err
	}

	// Move to the original path
	err = RenameAndReplace(tmpFile, dst)
	if err != nil {
		os.Remove(tmpFile)
		return err
	}

	return nil
}

func ReEncryptSingleAssetFile(src string, dst string, tmpFile string, generalProgress int64, current int, total int, srcKey []byte, dstKey []byte) (int64, error) {
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

	// Create read stream

	readStream, err := CreateFileBlockEncryptReadStream(src, srcKey)

	if err != nil {
		return 0, err
	}

	// Create and initialize write stream

	writeStream, err := CreateFileBlockEncryptWriteStream(tmpFile)

	if err != nil {
		readStream.Close()
		return 0, err
	}

	err = writeStream.Initialize(readStream.file_size, dstKey)

	if err != nil {
		readStream.Close()
		writeStream.Close()
		return 0, err
	}

	// Pipe

	buf := make([]byte, 32*1000)

	written := int64(0)
	progressInt := int64(0)
	prevProgress := int64(0)
	size := readStream.file_size

	if size < 1 {
		size = 1
	}

	for {
		nr, er := readStream.Read(buf)
		if nr > 0 {
			ew := writeStream.Write(buf[0:nr])
			written += int64(nr)
			if ew != nil {
				err = ew
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

	readStream.Close()

	if err != nil {
		writeStream.Close()
		os.Remove(tmpFile)
		return 0, err
	}

	err = writeStream.Close()

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

func ReEncryptMultiAssetFile(src string, dst string, tmpFile string, generalProgress int64, current int, total int, srcKey []byte, dstKey []byte) (int64, error) {
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

	// Create read stream
	readStream, err := CreateMultiFilePackReadStream(src)

	if err != nil {
		return 0, err
	}

	// Create and initialize write stream

	writeStream, err := CreateMultiFilePackWriteStream(tmpFile)

	if err != nil {
		readStream.Close()
		return 0, err
	}

	err = writeStream.Initialize(readStream.file_count)

	if err != nil {
		readStream.Close()
		writeStream.Close()
		return 0, err
	}

	// Pipe

	written := int64(0)
	progressInt := int64(0)
	prevProgress := int64(0)
	size := readStream.file_count

	if size < 1 {
		size = 1
	}

	for i := int64(0); i < readStream.file_count; i++ {
		// Read embedded file
		b, err := readStream.GetFile(i)

		if err != nil {
			readStream.Close()
			writeStream.Close()
			return 0, err
		}

		// Decrypt
		decrypted, err := decryptFileContents(b, srcKey)

		if err != nil {
			readStream.Close()
			writeStream.Close()
			return 0, err
		}

		// Re-Encrypt

		encrypted, err := encryptFileContents(decrypted, AES256_ZIP, dstKey)

		if err != nil {
			readStream.Close()
			writeStream.Close()
			return 0, err
		}

		// Store

		err = writeStream.PutFile(encrypted)

		if err != nil {
			readStream.Close()
			writeStream.Close()
			return 0, err
		}

		written++

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

	readStream.Close()

	if err != nil {
		writeStream.Close()
		os.Remove(tmpFile)
		return 0, err
	}

	err = writeStream.Close()

	if err != nil {
		os.Remove(tmpFile)
		return 0, err
	}

	err = RenameAndReplace(tmpFile, dst)

	if err != nil {
		os.Remove(tmpFile)
		return 0, err
	}

	return sourceFileStat.Size(), nil
}
