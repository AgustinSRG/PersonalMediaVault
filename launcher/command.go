// User commands management

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func readNextCommand(reader *bufio.Reader, vaultPathAbs string, vc *VaultController) {
	fmt.Print("PMV[" + vaultPathAbs + "]> ")

	ans, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
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
				openBrowser(vc.launchConfig.Port, vc.launchConfig.hasSSL())
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
				openBrowser(vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "browser", "b":
		openBrowser(vc.launchConfig.Port, vc.launchConfig.hasSSL())
	case "port", "p":
		if len(args) == 1 {
			fmt.Println("Listening port: " + fmt.Sprint(vc.launchConfig.Port))
		} else if len(args) == 2 {
			p, err := strconv.Atoi(args[1])

			if err != nil || p <= 0 {
				fmt.Println("Usage: port [p] - Sets the listening port")
			} else {
				vc.launchConfig.Port = p
				err := writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

				if err != nil {
					fmt.Println("Error: " + err.Error())
				} else {
					fmt.Println("Listening port changed: " + fmt.Sprint(vc.launchConfig.Port))
					askRestart(vc)
				}
			}
		} else {
			fmt.Println("Usage: port [p] - Sets the listening port")
		}
	case "local", "localhost", "l":
		if len(args) == 1 {
			if vc.launchConfig.Local {
				fmt.Println("Listening mode: Local (localhost)")
			} else {
				fmt.Println("Listening mode: All interfaces ([::])")
			}
		} else if len(args) == 2 {
			if strings.HasPrefix(strings.ToLower(args[1]), "y") {
				vc.launchConfig.Local = true
				err := writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

				if err != nil {
					fmt.Println("Error: " + err.Error())
				} else {
					fmt.Println("Listening mode changed: Local (localhost)")
					askRestart(vc)
				}
			} else if strings.HasPrefix(strings.ToLower(args[1]), "n") {
				vc.launchConfig.Local = false
				err := writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

				if err != nil {
					fmt.Println("Error: " + err.Error())
				} else {
					fmt.Println("Listening mode changed: All interfaces ([::])")
					askRestart(vc)
				}
			} else {
				fmt.Println("Usage: local [y/n] - Sets local listening mode")
			}
		} else {
			fmt.Println("Usage: local [y/n] - Sets local listening mode")
		}
	case "clean", "c":
		if vc.Stop() {
			vc.WaitForStop()
		}
		vc.Clean()
		if vc.Start() {
			if vc.WaitForStart() {
				openBrowser(vc.launchConfig.Port, vc.launchConfig.hasSSL())
			}
		}
	case "backup", "bkp", "bk":
		if len(args) == 2 {
			ap, err := filepath.Abs(args[1])
			if err != nil {
				fmt.Println("Usage: backup [path] - Makes a backup of the vault in the specified path")
			} else {
				vc.Backup(ap)
			}
		} else {
			fmt.Println("Usage: backup [path] - Makes a backup of the vault in the specified path")
		}
	case "ssl-check", "ssl":
		if vc.launchConfig.SSL_Cert == "" {
			fmt.Println("SSL certificate file: " + "(Not Set)")
		} else {
			fmt.Println("SSL certificate file: " + vc.launchConfig.SSL_Cert)
		}
		if vc.launchConfig.SSL_Key == "" {
			fmt.Println("SSL key file: " + "(Not Set)")
		} else {
			fmt.Println("SSL key file: " + vc.launchConfig.SSL_Key)
		}
	case "ssl-setup":
		if vc.SetupSSL() {
			askRestart(vc)
		}
	case "ssl-disable":
		if vc.disableSSL() {
			askRestart(vc)
		}
	case "secdel", "secure-delete":
		if len(args) == 2 {
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
				fmt.Println("Usage: secdel [y/n]")
			}
		} else {
			fmt.Println("Usage: secdel [y/n]")
		}
	case "help", "h", "commands", "man", "?":
		printCommandList()
	case "exit", "quit", "q":
		if vc.Stop() {
			vc.WaitForStop()
		}
		os.Exit(0)
	default:
		fmt.Println("Unrecognized command: '" + cmdKey + "'. Use 'help' to get the command list.")
	}
}

func askRestart(vc *VaultController) {
	fmt.Print("Restart the vault? (y/n): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if strings.HasPrefix(strings.ToLower(ans), "y") {
		runCommand("restart", vc)
	}
}

func printCommandList() {
	fmt.Println("    help          - Prints command list")
	fmt.Println("    exit          - Closes the vault and exits the program")
	fmt.Println("    start         - Starts the vault")
	fmt.Println("    stop          - Stops the vault")
	fmt.Println("    restart       - Restarts the vault")
	fmt.Println("    browser       - Opens the vault using the default browser")
	fmt.Println("    status        - Prints current status and configuration")
	fmt.Println("    clean         - Restarts the vault and cleans inconsistent files")
	fmt.Println("    port [p]      - Sets the listening port")
	fmt.Println("    local [y/n]   - Sets local listening mode")
	fmt.Println("    ssl           - Prints ssl configuration (if any)")
	fmt.Println("    ssl-setup     - Setups SSL to use HTTPS for accessing your vault")
	fmt.Println("    ssl-disable   - Disables SSL (use regular HTTP)")
	fmt.Println("    secdel [y/n]  - Enables / disables secure deletion of temp files")
	fmt.Println("    backup [path] - Makes a backup of the vault in the specified path")
}
