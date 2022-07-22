// User commands management

package main

import (
	"bufio"
	"fmt"
	"os"
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
				openBrowser(vc.launchConfig.Port)
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
			vc.WaitForStart()
		}
	case "browser", "b":
		openBrowser(vc.launchConfig.Port)
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

func printCommandList() {
	fmt.Println("    help    - Prints command list")
	fmt.Println("    exit    - Closes the vault and exits the program")
	fmt.Println("    start   - Starts the vault")
	fmt.Println("    stop    - Stops the vault")
	fmt.Println("    restart - Restarts the vault")
	fmt.Println("    status  - Prints current status and configuration")
	fmt.Println("    browser - Opens the vault using the default browser")
}
