// Key recovery

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

func runKeyRecoverCommand() {
	// Parse arguments
	args := os.Args

	if len(args) != 3 {
		printUsageKeyRecover()
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

	credentialFile := path.Join(vaultPath, "credentials.json")

	if !CheckFileExists(credentialFile) {
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
			ID:    "ModeKeyRecover",
			Other: "Mode: Recover encryption key",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	// Check lock to prevent recover if vault is being used

	if CheckFileExists(path.Join(vaultPath, "vault.lock")) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LockError",
				Other: "The destination path has a lock file. Close the vault or remove the vault.lock file.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Load credentials

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "LoadingCredentials",
			Other: "Loading credentials...",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	sourceCredentials, err := ReadVaultCredentials(credentialFile)

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

	// Ask for key

	fmt.Fprintln(os.Stderr, "")

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultKeyAsk1",
			Other: "Input the vault key, in hexadecimal format, and press enter:",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	fmt.Fprintln(os.Stderr, "")

	reader := bufio.NewReader(os.Stdin)

	vaultKeyHex, err := reader.ReadString('\n')

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

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Redacted",
			Other: "[REDACTED]",
		},
	})

	redactedMsg := msg

	for len(redactedMsg) < len(vaultKeyHex) {
		redactedMsg += " "
	}

	fmt.Fprintln(os.Stderr, "\033[F"+redactedMsg)
	fmt.Fprintln(os.Stderr, "")

	// Check key

	vaultKey, err := hex.DecodeString(strings.TrimSpace(vaultKeyHex))

	if err != nil {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultKeyInvalid",
				Other: "Error: Invalid vault key. {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Fprintln(os.Stderr, msg)

		os.Exit(1)
	}

	if len(vaultKey) != 32 {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultKeyInvalid",
				Other: "Error: Invalid vault key. Must be a 32 bytes hexadecimal string.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)

		os.Exit(1)
	}

	checkKeyForRecovery(vaultPath, vaultKey)

	// Final confirmation

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RecoverKeyFinalWarning",
			Other: "WARNING: This will clear all your vault accounts except the root one",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RecoverKeyFinalWarningProceed",
			Other: "Proceed with key recovery?",
		},
	})
	ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err := reader.ReadString('\n')
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

	ans = strings.TrimSpace(ans)

	if !checkYesNoAnswer(ans) {
		return
	}

	// Ask for new password

	var password string = ""
	var password_repeat string = ""

	for password == "" || password != password_repeat {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EnterRootPassword",
				Other: "Enter new password for the root account of the vault ({{.Account}})",
			},
			TemplateData: map[string]interface{}{
				"Account": sourceCredentials.User,
			},
		})
		fmt.Print(msg + ": ")
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

		password = strings.TrimSpace(string(bytePassword))

		if password == "" {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorPasswordBlank",
					Other: "Password cannot be blank.",
				},
			})
			fmt.Fprintln(os.Stderr, msg)
			continue
		}

		if len(password) > 255 {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorPasswordTooLong",
					Other: "Password cannot be longer than 255 characters.",
				},
			})
			fmt.Fprintln(os.Stderr, msg)
			password = ""
			continue
		}

		fmt.Print("\n")

		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "RepeatPassword",
				Other: "Repeat Password",
			},
		})
		fmt.Print(msg + ": ")
		bytePassword, err = term.ReadPassword(int(syscall.Stdin))
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

		fmt.Print("\n")

		password_repeat = strings.TrimSpace(string(bytePassword))

		if password != password_repeat {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorPasswordNotMatching",
					Other: "Passwords do not match.",
				},
			})
			fmt.Fprintln(os.Stderr, msg)
		}
	}

	// Recover the key

	err = sourceCredentials.RecoverKey(vaultKey, password)

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

	tempPath := path.Join(vaultPath, "temp")

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
	}

	tmpFile := path.Join(tempPath, "credentials.json.tmp")

	err = sourceCredentials.WriteToFile(credentialFile, tmpFile)

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

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RecoverKeyDone",
			Other: "Done! You may now login to the vault with your new password.",
		},
	})
	fmt.Fprintln(os.Stderr, msg)
}
