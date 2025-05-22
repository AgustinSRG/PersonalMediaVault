// Backup tool for PMV

package main

import (
	"os"
)

func main() {
	// Initialize
	InitializeInternationalizationFramework()

	// Parse arguments
	args := os.Args

	if len(args) < 2 {
		printHelp()
		return
	}

	command := args[1]

	if command == "backup" {
		runBackupCommand(false)
	} else if command == "re-encrypt" {
		runBackupCommand(true)
	} else if command == "key-export" {
		runKeyExportCommand()
	} else if command == "key-recover" {
		runKeyRecoverCommand()
	} else {
		printHelp()
		os.Exit(1)
	}
}
