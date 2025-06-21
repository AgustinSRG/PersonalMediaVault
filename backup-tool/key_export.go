// Key export

package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/term"
)

func runKeyExportCommand() {
	// Parse arguments
	args := os.Args

	if len(args) != 3 {
		printUsageKeyExport()
		os.Exit(1)
	}

	vaultPath := args[2]

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

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ModeKeyExport",
			Other: "Mode: Export encryption key",
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

	sourceCredentials, err := ReadVaultCredentials(path.Join(vaultPath, "credentials.json"))

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

	sourceVaultKeyHex := strings.ToUpper(hex.EncodeToString(sourceVaultKey))

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Redacted",
			Other: "[REDACTED]",
		},
	})

	redactedMsg := msg

	for len(redactedMsg) < len(sourceVaultKeyHex) {
		redactedMsg += " "
	}

	// Show key to console

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ExportKeyTitle1",
			Other: "Below is is the vault encryption key, in hexadecimal format.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ExportKeyTitle2",
			Other: "Make sure to store it securely. Never reveal it!",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ExportKeyTitle3",
			Other: "After you are done, press enter to clear the key from view.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	fmt.Fprintln(os.Stderr, "")

	fmt.Fprint(os.Stderr, sourceVaultKeyHex)

	reader := bufio.NewReader(os.Stdin)

	_, _ = reader.ReadString('\n')

	fmt.Fprintln(os.Stderr, "\033[F"+redactedMsg)
}
