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
}

// Program entry point
func main() {
	// Read arguments
	args := os.Args

	options := BackendOptions{
		debug:       false,
		ffmpegPath:  os.Getenv("FFMPEG_PATH"),
		ffprobePath: os.Getenv("FFPROBE_PATH"),
	}

	if options.ffmpegPath == "" {
		options.ffmpegPath = "/usr/bin/ffmpeg"
	}

	if options.ffprobePath == "" {
		options.ffprobePath = "/usr/bin/ffprobe"
	}

	for i := 1; i < (len(args) - 2); i++ {
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
			options.daemon = true
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
