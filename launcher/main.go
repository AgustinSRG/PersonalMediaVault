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
		vaultPath = "vault"
	} else {
		printHelp()
		os.Exit(1)
	}

	detectLauncherPaths()

	absolutePath, err := filepath.Abs(vaultPath)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	printVersion()

	fmt.Println("Vault path: " + absolutePath)

	reader := bufio.NewReader(os.Stdin)

	if !folderExists(vaultPath) {
		fmt.Print("Vault folder does not exists, do you want to create it? (y/n): ")
		ans, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if strings.HasPrefix(strings.ToLower(ans), "y") {
			err = os.MkdirAll(vaultPath, FOLDER_PERMISSION)
			if err != nil {
				fmt.Println("Error: " + err.Error())
				os.Exit(1)
			}
		} else {
			return
		}
	}

	launcherConfigFile := path.Join(vaultPath, "launcher.config.json")

	launcherConfig := readLauncherConfig(launcherConfigFile)

	for launcherConfig.Port <= 0 {
		fmt.Println("Please, choose a port for the backend to listen.")
		fmt.Print("Port number [80]: ")

		ans, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if ans == "" {
			ans = "80"
		}

		p, err := strconv.ParseInt(ans, 10, 64)

		launcherConfig.Port = int(p)

		if launcherConfig.Port <= 0 {
			continue
		}

		fmt.Print("Do you want to bind to localhost? (y/n) (by selecting no, it will bind all network interfaces): ")

		ans, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if strings.HasPrefix(strings.ToLower(ans), "y") {
			launcherConfig.Local = true
		} else {
			launcherConfig.Local = false
		}

		err = writeLauncherConfig(launcherConfigFile, launcherConfig)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}
	}

	if launcherConfig.Local {
		fmt.Println("Configured listening address as localhost:" + fmt.Sprint(launcherConfig.Port))
	} else {
		fmt.Println("Configured listening address as [::]:" + fmt.Sprint(launcherConfig.Port))
	}

	if fileExists(path.Join(vaultPath, "vault.lock")) {
		fmt.Println("Seems like the vault is being used by another process.")
		fmt.Println("Openning the vault by multiple processes could be dangerous for the vault integrity.")
		fmt.Print("Procceed? (y/n): ")

		ans, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if strings.HasPrefix(strings.ToLower(ans), "y") {
			os.Remove(path.Join(vaultPath, "vault.lock"))
		} else {
			return
		}
	}

	var vaultController VaultController

	// Initailize the vault if needed

	vaultController.Initialize(absolutePath, launcherConfig, reader)

	// Start vault

	runCommand("start", &vaultController)

	// Read commands

	for true {
		readNextCommand(reader, absolutePath, &vaultController)
	}
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
