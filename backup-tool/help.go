// Help texts

package main

import (
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func printUsageBackup() {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageBackup",
			Other: "Usage: pmv-backup backup </path/to/vault> </path/to/backup>",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}

func printUsageReEncrypt() {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageReEncrypt",
			Other: "Usage: pmv-backup re-encrypt </path/to/vault> </path/to/backup>",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}

func printHelp() {
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
			Other: "Usage: pmv-backup <command> [arguments]",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageBackupTitle",
			Other: "If you want to create or sync a backup, run the 'backup' command.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
	printUsageBackup()

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageReEncryptTitle",
			Other: "If you want to re-encrypt the vault with a different encryption key, run the 're-encrypt' command.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
	printUsageReEncrypt()
}
