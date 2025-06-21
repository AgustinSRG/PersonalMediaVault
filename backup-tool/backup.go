// Backup

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Runs the backup or re-encrypt command
func runBackupCommand(isReEncrypt bool) {
	// Parse arguments
	args := os.Args

	if len(args) != 4 {
		if isReEncrypt {
			printUsageReEncrypt()
		} else {
			printUsageBackup()
		}
		os.Exit(1)
	}

	vaultPath := args[2]
	backupPath := args[3]

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

	// Create temp path for atomic copy

	tempPath := path.Join(backupPath, "temp")

	err = os.MkdirAll(tempPath, FOLDER_PERMISSION)

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
		return
	}

	tmpFile := path.Join(tempPath, "backup.tmp")

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

	if isReEncrypt {
		RunReEncryptedBackup(vaultPath, backupPath, tmpFile)
	} else {
		RunNormalBackup(vaultPath, backupPath, tmpFile)
	}
}
