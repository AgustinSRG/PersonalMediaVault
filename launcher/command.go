// User commands management

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rivo/uniseg"
	"golang.org/x/term"
)

func readNextCommand(reader *bufio.Reader, vaultPathAbs string, vc *VaultController) {
	fmt.Print("PMV[" + vaultPathAbs + "]> ")

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

	runCommand(ans, vc)
}

func runCommand(cmdText string, vc *VaultController) {
	if cmdText == "" {
		return
	}

	args := strings.Fields(cmdText)

	cmdKey := strings.ToLower(args[0])

	switch cmdKey {
	case "start", "up":
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "stop", "down":
		if vc.Stop() {
			vc.WaitForStop()
		}
	case "status", "check", "s":
		vc.PrintStatus()
	case "restart", "rs":
		if vc.Stop() {
			vc.WaitForStop()
		}
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "browser", "b":
		openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
	case "open-logs", "logs":
		logsPath, err := getLogsPath()

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
			return
		}

		openFileExplorer(logsPath)
	case "host", "hostname":
		if len(args) == 1 {
			hostName := vc.launchConfig.HostName

			if hostName == "" {
				hostName = "localhost"
			}

			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "VaultHostname",
					Other: "Vault host: {{.HostName}}",
				},
				TemplateData: map[string]interface{}{
					"HostName": hostName,
				},
			})
			fmt.Println(msg)
		} else if len(args) == 2 {
			r, err := regexp.Compile("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$")

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
				return
			}

			hostName := args[1]

			if !r.MatchString(hostName) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorHostNameInvalid",
						Other: "Error: Invalid hostname provided",
					},
				})
				fmt.Println(msg)
				return
			}

			if vc.SetHostName(hostName) {
				askRestart(vc)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorHostUsage",
					Other: "Usage: host [hostname] - Sets the vault hostname for the browser",
				},
			})
			fmt.Println(msg)
		}
	case "port", "p":
		if len(args) == 1 {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ListeningPort",
					Other: "Listening port: {{.Port}}",
				},
				TemplateData: map[string]interface{}{
					"Port": fmt.Sprint(vc.launchConfig.Port),
				},
			})
			fmt.Println(msg)
		} else if len(args) == 2 {
			p, err := strconv.Atoi(args[1])

			if err != nil || p <= 0 {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorPortUsage",
						Other: "Usage: port [p] - Sets the listening port",
					},
				})
				fmt.Println(msg)
			} else {
				vc.launchConfig.Port = p
				err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

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
				} else {
					msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "ListeningPortChanged",
							Other: "Listening port changed: {{.Port}}",
						},
						TemplateData: map[string]interface{}{
							"Port": fmt.Sprint(vc.launchConfig.Port),
						},
					})
					fmt.Println(msg)
					askRestart(vc)
				}
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorPortUsage",
					Other: "Usage: port [p] - Sets the listening port",
				},
			})
			fmt.Println(msg)
		}
	case "local", "localhost", "l":
		if len(args) == 1 {
			if vc.launchConfig.Local {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ListeningModeLocal",
						Other: "Listening mode: Local",
					},
				})
				fmt.Println(msg + " (localhost)")
			} else {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ListeningModeAll",
						Other: "Listening mode: All interfaces",
					},
				})
				fmt.Println(msg + " ([::])")
			}
		} else if len(args) == 2 {
			if checkYesNoAnswer(args[1]) {
				vc.launchConfig.Local = true
				err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

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
				} else {
					msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "ListeningModeLocal",
							Other: "Listening mode: Local",
						},
					})
					fmt.Println(msg + " (localhost)")
					askRestart(vc)
				}
			} else if checkNegativeAnswer(args[1]) {
				vc.launchConfig.Local = false
				err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

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
				} else {
					msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "ListeningModeAll",
							Other: "Listening mode: All interfaces",
						},
					})
					fmt.Println(msg + " ([::])")
					askRestart(vc)
				}
			} else {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorLocalUsage",
						Other: "Usage: local [y/n] - Sets local listening mode",
					},
				})
				fmt.Println(msg)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorLocalUsage",
					Other: "Usage: local [y/n] - Sets local listening mode",
				},
			})
			fmt.Println(msg)
		}
	case "clean":
		if vc.Stop() {
			vc.WaitForStop()
		}
		vc.Clean()
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "recover":
		if vc.Stop() {
			vc.WaitForStop()
		}
		vc.RecoverAssets()
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "backup", "bkp", "bk":
		if len(args) == 2 {
			ap, err := filepath.Abs(args[1])
			if err != nil {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorBackupUsage",
						Other: "Usage: backup [path] - Makes a backup of the vault in the specified path",
					},
				})
				fmt.Println(msg)
			} else {
				vc.Backup(ap, false)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorBackupUsage",
					Other: "Usage: backup [path] - Makes a backup of the vault in the specified path",
				},
			})
			fmt.Println(msg)
		}
	case "re-encrypt":
		if len(args) == 2 {
			ap, err := filepath.Abs(args[1])
			if err != nil {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorReEncryptUsage",
						Other: "Usage: re-encrypt [path] - Makes a re-encrypted backup of the vault in the specified path",
					},
				})
				fmt.Println(msg)
			} else {
				vc.Backup(ap, true)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorReEncryptUsage",
					Other: "Usage: re-encrypt [path] - Makes a re-encrypted backup of the vault in the specified path",
				},
			})
			fmt.Println(msg)
		}
	case "key-export":
		vc.KeyExport()
	case "key-recover":
		if vc.Stop() {
			vc.WaitForStop()
		}
		vc.KeyRecover()
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.HostName, vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "ssl-check", "ssl":
		if vc.launchConfig.SSL_Cert == "" {
			notSetMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "NotSet",
					Other: "(Not Set)",
				},
			})
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SSLCertFile",
					Other: "SSL certificate file: {{.File}}",
				},
				TemplateData: map[string]interface{}{
					"File": notSetMsg,
				},
			})
			fmt.Println(msg)
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SSLCertFile",
					Other: "SSL certificate file: {{.File}}",
				},
				TemplateData: map[string]interface{}{
					"File": vc.launchConfig.SSL_Cert,
				},
			})
			fmt.Println(msg)
		}
		if vc.launchConfig.SSL_Key == "" {
			notSetMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "NotSet",
					Other: "(Not Set)",
				},
			})
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SSLkeyFile",
					Other: "SSL key file: {{.File}}",
				},
				TemplateData: map[string]interface{}{
					"File": notSetMsg,
				},
			})
			fmt.Println(msg)
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SSLkeyFile",
					Other: "SSL key file: {{.File}}",
				},
				TemplateData: map[string]interface{}{
					"File": vc.launchConfig.SSL_Key,
				},
			})
			fmt.Println(msg)
		}
	case "ssl-setup":
		if vc.SetupSSL() {
			askRestart(vc)
		}
	case "ssl-disable":
		if vc.disableSSL() {
			askRestart(vc)
		}
	case "sec-del", "secdel", "secure-delete":
		if len(args) == 2 {
			if checkYesNoAnswer(args[1]) {
				args[1] = "y"
			}
			switch args[1] {
			case "y", "on", "yes", "1", "true":
				if vc.SetSecureTempDelete(true) {
					askRestart(vc)
				}
			case "n", "off", "no", "0", "false":
				if vc.SetSecureTempDelete(false) {
					askRestart(vc)
				}
			default:
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorSecDelUsage",
						Other: "Usage: sec-del [y/n]",
					},
				})
				fmt.Println(msg)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorSecDelUsage",
					Other: "Usage: sec-del [y/n]",
				},
			})
			fmt.Println(msg)
		}
	case "log-requests", "requests-log":
		if len(args) == 2 {
			if checkYesNoAnswer(args[1]) {
				args[1] = "y"
			}
			switch args[1] {
			case "y", "on", "yes", "1", "true":
				if vc.SetLogRequests(true) {
					askRestart(vc)
				}
			case "n", "off", "no", "0", "false":
				if vc.SetLogRequests(false) {
					askRestart(vc)
				}
			default:
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorLogRequestsUsage",
						Other: "Usage: log-requests [y/n]",
					},
				})
				fmt.Println(msg)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorLogRequestsUsage",
					Other: "Usage: log-requests [y/n]",
				},
			})
			fmt.Println(msg)
		}
	case "debug", "log-debug", "debug-log":
		if len(args) == 2 {
			if checkYesNoAnswer(args[1]) {
				args[1] = "y"
			}
			switch args[1] {
			case "y", "on", "yes", "1", "true":
				if vc.SetLogDebug(true) {
					askRestart(vc)
				}
			case "n", "off", "no", "0", "false":
				if vc.SetLogDebug(false) {
					askRestart(vc)
				}
			default:
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorLogDebugUsage",
						Other: "Usage: debug [y/n]",
					},
				})
				fmt.Println(msg)
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorLogDebugUsage",
					Other: "Usage: debug [y/n]",
				},
			})
			fmt.Println(msg)
		}
	case "cache-size", "cs":
		if len(args) == 1 {
			currentCacheSize := DEFAULT_CACHE_SIZE

			if vc.launchConfig.CacheSize != nil {
				currentCacheSize = *(vc.launchConfig.CacheSize)
			}

			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CacheSize",
					Other: "Current cache size: {{.Elements}} elements",
				},
				TemplateData: map[string]interface{}{
					"Elements": fmt.Sprint(currentCacheSize),
				},
			})
			fmt.Println(msg)
		} else if len(args) == 2 {
			s, err := strconv.Atoi(args[1])

			if err != nil || s < 0 {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorCacheSizeUsage",
						Other: "Usage: cache-size [size] - Sets the cache size",
					},
				})
				fmt.Println(msg)
			} else {
				if vc.SetCacheSize(s) {
					askRestart(vc)
				}
			}
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorCacheSizeUsage",
					Other: "Usage: cache-size [size] - Sets the cache size",
				},
			})
			fmt.Println(msg)
		}
	case "help", "h", "commands", "man", "?":
		printCommandList()
	case "exit", "quit", "q":
		if vc.Stop() {
			vc.WaitForStop()
		}
		os.Exit(0)
	default:
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "UnrecognizedCommand",
				Other: "Unrecognized command: '{{.Command}}'. Use 'help' to get the command list.",
			},
			TemplateData: map[string]interface{}{
				"Command": cmdKey,
			},
		})
		fmt.Println(msg)
	}
}

func askRestart(vc *VaultController) {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RestartVaultAsk",
			Other: "Restart the vault?",
		},
	})
	ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err := vc.consoleReader.ReadString('\n')
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
		runCommand("restart", vc)
	}
}

func printCommandList() {
	manList := make([]string, 0)

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandHelp",
			Other: "help - Prints command list",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandExit",
			Other: "exit - Closes the vault and exits the program",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandStart",
			Other: "start - Starts the vault",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandStop",
			Other: "stop - Stops the vault",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandRestart",
			Other: "restart - Restarts the vault",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandBrowser",
			Other: "browser - Opens the vault using the default browser",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandStatus",
			Other: "status - Prints current status and configuration",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandClean",
			Other: "clean - Restarts the vault and cleans inconsistent files",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandRecover",
			Other: "recover - Restarts the vault and recovers any non-indexed media",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandPort",
			Other: "port [p] - Sets the listening port",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandHostname",
			Other: "host [hostname] - Sets the hostname to access the vault",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandLocal",
			Other: "local [y/n] - Sets local listening mode",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandSSL",
			Other: "ssl - Prints ssl configuration (if any)",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandSSLSetup",
			Other: "ssl-setup - Setups SSL to use HTTPS for accessing your vault",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandSSLDisable",
			Other: "ssl-disable - Disables SSL (use regular HTTP)",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandSecDel",
			Other: "sec-del [y/n] - Enables / disables secure deletion of temp files",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandOpenLogs",
			Other: "open-logs - Opens the logs folder in the file explorer",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandLogRequests",
			Other: "log-requests [y/n] - Enables / disables request logging",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandLogDebug",
			Other: "debug [y/n] - Enables / disables debug logging",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCacheSize",
			Other: "cache-size [size] - Sets the cache size",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandBackup",
			Other: "backup [path] - Makes a backup of the vault in the specified path",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandReEncrypt",
			Other: "re-encrypt [path] - Makes a re-encrypted backup of the vault in the specified path",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandKeyExport",
			Other: "key-export - Exports the encryption key of the vault, in order to make a backup",
		},
	})
	manList = append(manList, msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ManualCommandKeyRecover",
			Other: "key-recover - Recovers access to the vault, using a backup of the encryption key",
		},
	})
	manList = append(manList, msg)

	fmt.Println(prepareCommandManualList(manList))
}

func prepareCommandManualList(manList []string) string {
	// Check the largest key

	largestKeyLength := 0

	for i := 0; i < len(manList); i++ {
		parts := strings.Split(manList[i], " - ")

		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])

		if uniseg.GraphemeClusterCount(key) > largestKeyLength {
			largestKeyLength = uniseg.GraphemeClusterCount(key)
		}
	}

	// Prepare string

	result := ""
	first := true

	termSize, _, err := term.GetSize(int(syscall.Stdout))

	termSize -= 10 // Padding

	if err != nil || termSize < 80 {
		termSize = 80 // Min allowed size: 80 chars
	}

	for i := 0; i < len(manList); i++ {
		parts := strings.Split(manList[i], " - ")

		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])

		for uniseg.GraphemeClusterCount(key) < largestKeyLength {
			key = key + " "
		}

		if !first {
			result += "\n"
		} else {
			first = false
		}

		desc := strings.TrimSpace(strings.Join(parts[1:], " - "))

		linePrefix := "    " + key + " - "

		linePrefixSpaces := ""

		for j := 0; j < uniseg.GraphemeClusterCount(linePrefix); j++ {
			linePrefixSpaces += " "
		}

		sizeAvailableForDescription := termSize - uniseg.GraphemeClusterCount(linePrefix)

		descLines := make([]string, 0)
		curDescLine := ""

		curDescLineState := -1
		var curDescWord string

		for len(desc) > 0 {
			curDescWord, desc, curDescLineState = uniseg.FirstWordInString(desc, curDescLineState)

			if len(curDescLine) > 0 && uniseg.GraphemeClusterCount(curDescLine+curDescWord) > sizeAvailableForDescription {
				descLines = append(descLines, curDescLine)
				curDescLine = ""
			}

			curDescLine += curDescWord
		}

		if len(curDescLine) > 0 {
			descLines = append(descLines, curDescLine)
		}

		if len(descLines) == 0 {
			continue
		}

		result += linePrefix + descLines[0]

		for j := 1; j < len(descLines); j++ {
			result += "\n" + linePrefixSpaces + strings.TrimSpace(descLines[j])
		}
	}

	return result
}
