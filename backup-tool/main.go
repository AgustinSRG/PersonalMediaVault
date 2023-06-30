// Backup tool for PMV

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func main() {
	// Initialize
	InitializeInternationalizationFramework()

	// Parse arguments
	args := os.Args

	if len(args) != 3 {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ProgramWelcome",
				Other: "Backup tool for Personal Media Vault.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ProgramUsage",
				Other: "Usage: pmv-backup </path/to/vault> </path/to/backup>",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		return
	}

	vaultPath := args[1]
	backupPath := args[2]

	if !CheckFileExists(vaultPath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PathError",
				Other: "Path does not exist: {{.Path}}",
			},
			TemplateData: map[string]interface{}{
				"Path": vaultPath,
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Create destination path if not found

	err := os.MkdirAll(backupPath, FOLDER_PERMISSION)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PathError",
				Other: "Path does not exist: {{.Path}}",
			},
			TemplateData: map[string]interface{}{
				"Path": backupPath,
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	if CheckFileExists(path.Join(backupPath, "vault.lock")) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LockError",
				Other: "The destination path has a lock file. Close the vault or remove the vault.lock file.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Welcome

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramWelcome",
			Other: "Backup tool for Personal Media Vault.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	// Fetch vault metadata

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "FetchNotice",
			Other: "Fetching metadata from: {{.Path}}",
		},
		TemplateData: map[string]interface{}{
			"Path": vaultPath,
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	if !CheckFileExists(path.Join(vaultPath, "credentials.json")) {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNotFoundError",
				Other: "Could not find a vault in the specified path",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Find files to check

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

		copied, err := backupFile(work.entries[i], progressInt, i+1, len(work.entries))

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
