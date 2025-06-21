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

func printUsageKeyExport() {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageKeyExport",
			Other: "Usage: pmv-backup key-export </path/to/vault>",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}

func printUsageKeyRecover() {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageKeyRecover",
			Other: "Usage: pmv-backup key-recover </path/to/vault>",
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

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageKeyExportTitle",
			Other: "If you want to export the encryption key in order to save a backup of it, run the 'key-export' command.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
	printUsageKeyExport()

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ProgramUsageKeyRecoverTitle",
			Other: "If you lost access to the vault and you have the encryption key, you can run the 'key-recover' command to restore access.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
	printUsageKeyRecover()
}
