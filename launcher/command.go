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
	case "start":
		if vc.Start() {
			vc.WaitForStart()
			openBrowser(vc.launchConfig.Port)
		}
	case "stop":
		if vc.Stop() {
			vc.WaitForStop()
		}
	case "browser":
		openBrowser(vc.launchConfig.Port)
	case "exit", "quit", "q":
		if vc.Stop() {
			vc.WaitForStop()
		}
		os.Exit(0)
	}
}
