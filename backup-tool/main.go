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
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ProgramUsage",
				Other: "Usage: pmv-backup </path/to/vault> </path/to/backup>",
			},
		})
		fmt.Println(msg)
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
		fmt.Println(msg)
		os.Exit(1)
	}

	if !CheckFileExists(backupPath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PathError",
				Other: "Path does not exist: {{.Path}}",
			},
			TemplateData: map[string]interface{}{
				"Path": backupPath,
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramWelcome",
			Other: "Backup tool for Personal Media Vault.",
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "FetchNotice",
			Other: "Fetching metadata from: {{.Path}}",
		},
		TemplateData: map[string]interface{}{
			"Path": vaultPath,
		},
	})
	fmt.Println(msg)

	if !CheckFileExists(path.Join(vaultPath, "credentials.json")) {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNotFoundError",
				Other: "Could not find a vault in the specified path",
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	totalEntries := make([]BackupEntry, 0)

	mediaFiles := findBackupEntries(vaultPath, backupPath, "./media")
	totalEntries = append(totalEntries, mediaFiles...)

	tagFiles := findBackupEntries(vaultPath, backupPath, "./tags")
	totalEntries = append(totalEntries, tagFiles...)

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "main.index"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "credentials.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "media_ids.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tasks.json"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "albums.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tag_list.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "user_config.pmv"))

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InitializeNotice",
			Other: "Initializing backup...",
		},
	})
	fmt.Println(msg)

	progressInt := int64(0)
	prevProgress := int64(0)

	statFilesCopied := 0

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "MakeNotice",
			Other: "Making backup...",
		},
	})
	fmt.Println(msg)

	for i := 0; i < len(totalEntries); i++ {

		copied, err := backupFile(totalEntries[i])

		if err != nil {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "BackupFailed",
					Other: "Backup failed!",
				},
			})
			fmt.Println(msg)
			os.Exit(1)
			return
		}

		if copied {
			statFilesCopied++
		}

		progressInt = int64(i+1) * 100 / int64(len(totalEntries))
		if prevProgress != progressInt {
			prevProgress = progressInt
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "BackupProgress",
					Other: "Making backup... ({{.Percent}}%)",
				},
				TemplateData: map[string]interface{}{
					"Percent": fmt.Sprint(progressInt),
				},
			})
			fmt.Print("\r" + msg)
		}
	}

	fmt.Print("\n")
	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "BackupDone",
			Other: "Backup done. Total files copied: {{.FileCount}}",
		},
		TemplateData: map[string]interface{}{
			"FileCount": fmt.Sprint(statFilesCopied),
		},
	})
	fmt.Println(msg)
}
