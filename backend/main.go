// Main

package main

import (
	"fmt"
	"os"
)

type BackendOptions struct {
	debug bool // Debug mode

	// Run modes
	daemon bool

	// FFmpeg
	ffmpegPath  string
	ffprobePath string

	// Vault
	vaultPath string

	// Temp path
	tempPath string
}

// Program entry point
func main() {
	// Read arguments
	args := os.Args

	options := BackendOptions{
		ffmpegPath:  os.Getenv("FFMPEG_PATH"),
		ffprobePath: os.Getenv("FFPROBE_PATH"),
		vaultPath:   "./vault",
		tempPath:    os.Getenv("TEMP_PATH"),
	}

	if options.ffmpegPath == "" {
		options.ffmpegPath = "/usr/bin/ffmpeg"
	}

	if options.ffprobePath == "" {
		options.ffprobePath = "/usr/bin/ffprobe"
	}

	if options.tempPath == "" {
		options.tempPath = "./temp"
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]

		if arg == "--debug" {
			options.debug = true
		} else if arg == "--help" || arg == "-h" {
			printHelp()
			return
		} else if arg == "--version" || arg == "-v" {
			printVersion()
			return
		} else if arg == "--daemon" || arg == "-d" {
			options.daemon = true
		} else if arg == "--vault-path" || arg == "-vp" {
			if i == len(args)-1 {
				fmt.Println("The option '--vault-path' requires a value")
				options.vaultPath = args[i+1]
				i++
			}
		} else {
			fmt.Println("Invalid argument: " + arg)
			os.Exit(1)
		}
	}

	if options.daemon {
		if _, err := os.Stat(options.ffmpegPath); err != nil {
			fmt.Println("Error: Could not find 'ffmpeg' at specified location: " + options.ffmpegPath)
			os.Exit(1)
		}

		if _, err := os.Stat(options.ffprobePath); err != nil {
			fmt.Println("Error: Could not find 'ffprobe' at specified location: " + options.ffprobePath)
			os.Exit(1)
		}

		SetFFMPEGBinaries(options.ffmpegPath, options.ffprobePath) // Set FFMPEG paths

		SetDebugLogEnabled(options.debug) // Log debug mode

		SetTempFilesPath(options.tempPath) // Set temporal files path

		// Create and initialize vault

		vault := Vault{}
		err := vault.Initialize(options.vaultPath)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		// TODO: Create and run HTTP server
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage: personal-media-vault [OPTIONS]")
	fmt.Println("    OPTIONS:")
	fmt.Println("        --help, -h                 Prints command line options.")
	fmt.Println("        --version, -v              Prints version.")
	fmt.Println("        --daemon, -d               Runs backend daemon.")
	fmt.Println("        --debug                    Enables debug mode.")
	fmt.Println("        --vault-path, -vp <path>   Sets the data storage path for the vault.")
	fmt.Println("    ENVIRONMENT VARIABLES:")
	fmt.Println("        FFMPEG_PATH                Path to ffmpeg binary.")
	fmt.Println("        FFPROBE_PATH               Path to ffprobe binary.")
	fmt.Println("        TEMP_PATH                  Path to a folder where temporally store files.")
}

func printVersion() {
	fmt.Println("PersonalMediaVault 1.0.0")
}
