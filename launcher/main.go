// Main

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const VERSION = "1.0.0"

// Program entry point
func main() {
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
		printHelp()
		return
	} else {
		printHelp()
		os.Exit(1)
	}

	absolutePath, err := filepath.Abs(vaultPath)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	printVersion()

	fmt.Println("Vault path: " + absolutePath)
}

func printHelp() {
	fmt.Println("Usage: pmv [PATH]")
	fmt.Println("Launches a vault with an interactive command line to manage it.")
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
	fmt.Println("- Personal Media Vault")
	fmt.Println("- Version " + VERSION)
	fmt.Println("- https://github.com/AgustinSRG/PersonalMediaVault")
	fmt.Println("---------------------------------------------------")
}
