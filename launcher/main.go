// Main

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const VERSION = "1.13.3"

// Program entry point
func main() {
	InitializeInternationalizationFramework()

	// Read arguments
	args := os.Args

	var vaultPath string

	if len(args) == 2 {
		vaultPath = args[1]

		if vaultPath == "-h" || vaultPath == "--help" {
			printHelp()
			return
		}

		if vaultPath == "-v" || vaultPath == "--version" {
			printVersion()
			return
		}
	} else if len(args) == 1 {
		// Default vaulty path
		userConfigDir, err := os.UserConfigDir()

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
			fmt.Println(msg)
			os.Exit(1)
		}

		vaultPath = path.Join(userConfigDir, "PersonalMediaVault", "vault")
	} else {
		printHelp()
		os.Exit(1)
	}

	err := child_process_manager.InitializeChildProcessManager()
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
		fmt.Println(msg)
		os.Exit(1)
	}
	defer child_process_manager.DisposeChildProcessManager() //nolint:errcheck

	detectLauncherPaths()

	vaultPath, err = filepath.Abs(vaultPath)

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
		fmt.Println(msg)
		os.Exit(1)
	}

	printVersion()

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultPath",
			Other: "Vault path: {{.Path}}",
		},
		TemplateData: map[string]interface{}{
			"Path": vaultPath,
		},
	})
	fmt.Println(msg)

	reader := bufio.NewReader(os.Stdin)

	if !folderExists(vaultPath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNotFoundCreateAsk",
				Other: "Vault folder does not exists, do you want to create it?",
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
			fmt.Println(msg)
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if checkYesNoAnswer(ans) {
			err = os.MkdirAll(vaultPath, FOLDER_PERMISSION)
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
				fmt.Println(msg)
				os.Exit(1)
			}
		} else {
			return
		}
	}

	launcherConfigFile := getLauncherConfigFile(vaultPath)

	launcherConfig := readLauncherConfig(launcherConfigFile)
	launcherConfig.Path = vaultPath

	for launcherConfig.Port <= 0 {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ChoosePort",
				Other: "Please, choose a port for the backend to listen.",
			},
		})
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PortNumber",
				Other: "Port number",
			},
		})
		fmt.Print(msg + " [80]: ")

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
			fmt.Println(msg)
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if ans == "" {
			ans = "80"
		}

		p, err := strconv.ParseInt(ans, 10, 64)

		if err != nil {
			continue
		}

		launcherConfig.Port = int(p)

		if launcherConfig.Port <= 0 {
			continue
		}

		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LocalHostBindAsk",
				Other: "Do you want to bind to localhost?",
			},
		})
		msg2, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LocalHostBindDesc",
				Other: "by selecting no, it will bind all network interface",
			},
		})
		ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YesNo",
				Other: "y/n",
			},
		})

		fmt.Print(msg + " (" + ynMsg + ") (" + msg2 + "): ")

		ans, err = reader.ReadString('\n')
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
			fmt.Println(msg)
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if ans == "" {
			ans = "y"
		}

		if checkYesNoAnswer(ans) {
			launcherConfig.Local = true
		} else {
			launcherConfig.Local = false
		}

		err = writeLauncherConfig(launcherConfigFile, launcherConfig)

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
			fmt.Println(msg)
			os.Exit(1)
		}
	}

	if launcherConfig.Local {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfiguredAddress",
				Other: "Configured listening address as {{.Address}}",
			},
			TemplateData: map[string]interface{}{
				"Address": "localhost:" + fmt.Sprint(launcherConfig.Port),
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfiguredAddress",
				Other: "Configured listening address as {{.Address}}",
			},
			TemplateData: map[string]interface{}{
				"Address": "[::]:" + fmt.Sprint(launcherConfig.Port),
			},
		})
		fmt.Println(msg)
	}

	if CheckVaultLocked(path.Join(vaultPath, "vault.lock")) {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultBeingUsed",
				Other: "Seems like the vault is being used by another process",
			},
		})
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "OpenMultipleRisk",
				Other: "Opening the vault by multiple processes could be dangerous for the vault integrity.",
			},
		})
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Proceed",
				Other: "Proceed?",
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
			fmt.Println(msg)
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if checkYesNoAnswer(ans) {
			os.Remove(path.Join(vaultPath, "vault.lock"))
		} else {
			return
		}
	}

	var vaultController VaultController

	// Initailize the vault if needed

	vaultController.Initialize(vaultPath, launcherConfig, reader)

	// Start vault

	runCommand("start", &vaultController)

	// Read commands

	for {
		readNextCommand(reader, vaultPath, &vaultController)
	}
}

func printHelp() {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Usage",
			Other: "Usage: pmv [PATH]",
		},
	})
	fmt.Println(msg)
	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Description",
			Other: "Launches a vault with an interactive command line to manage it.",
		},
	})
	fmt.Println(msg)
}

func printVersion() {
	fmt.Println("---------------------------------------------------")
	fmt.Println("-  _____    __  __  __      __")
	fmt.Println("- |  __ \\  |  \\/  | \\ \\    / /")
	fmt.Println("- | |__) | | \\  / |  \\ \\  / /")
	fmt.Println("- |  ___/  | |\\/| |   \\ \\/ /")
	fmt.Println("- | |      | |  | |    \\  /")
	fmt.Println("- |_|      |_|  |_|     \\/")
	fmt.Println("---------------------------------------------------")
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "PMV",
			Other: "Personal Media Vault",
		},
	})
	fmt.Println("- " + msg)
	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Version",
			Other: "Version",
		},
	})
	fmt.Println("- " + msg + " " + VERSION)
	fmt.Println("- https://github.com/AgustinSRG/PersonalMediaVault")
	fmt.Println("---------------------------------------------------")
}
