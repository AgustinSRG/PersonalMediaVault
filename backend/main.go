// Main

package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
)

type BackendOptions struct {
	debug bool // Debug mode

	// Run modes
	daemon     bool
	initialize bool
	clean      bool

	// FFmpeg
	ffmpegPath  string
	ffprobePath string

	// Vault
	vaultPath string

	// Temp path
	tempPath            string
	unencryptedTempPath string
}

var (
	GLOBAL_VAULT *Vault = nil
)

func GetVault() *Vault {
	return GLOBAL_VAULT
}

// Program entry point
func main() {
	// Read arguments
	args := os.Args

	options := BackendOptions{
		ffmpegPath:          os.Getenv("FFMPEG_PATH"),
		ffprobePath:         os.Getenv("FFPROBE_PATH"),
		vaultPath:           "./vault",
		tempPath:            os.Getenv("TEMP_PATH"),
		unencryptedTempPath: os.Getenv("UNENCRYPTED_TEMP_PATH"),
	}

	if options.ffmpegPath == "" {
		if runtime.GOOS == "windows" {
			options.ffmpegPath = "/ffmpeg/bin/ffmpeg.exe"
		} else {
			options.ffmpegPath = "/usr/bin/ffmpeg"
		}
	}

	if options.ffprobePath == "" {
		if runtime.GOOS == "windows" {
			options.ffprobePath = "/ffmpeg/bin/ffprobe.exe"
		} else {
			options.ffprobePath = "/usr/bin/ffprobe"
		}
	}

	if options.tempPath == "" {
		options.tempPath = "./temp"
	}

	if options.unencryptedTempPath == "" {
		userhome, err := os.UserHomeDir()

		if err != nil {
			LogError(err)
			os.Exit(1)
		}

		options.unencryptedTempPath = path.Join(userhome, ".pmv", "temp")
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
		} else if arg == "--clean" || arg == "-c" {
			options.clean = true
		} else if arg == "--cors-insecure" {
			CORS_INSECURE_MODE_ENABLED = true
		} else if arg == "--init" || arg == "-i" {
			options.initialize = true
		} else if arg == "--vault-path" || arg == "-vp" {
			if i == len(args)-1 {
				fmt.Println("The option '--vault-path' requires a value")
				os.Exit(1)
			}
			options.vaultPath = args[i+1]
			i++
		} else {
			fmt.Println("Invalid argument: " + arg)
			os.Exit(1)
		}
	}

	if options.daemon || options.clean || options.initialize {
		// Setup lockfile
		if !TryLockVault(options.vaultPath) {
			fmt.Println("Error: The vault is already in use.")
			os.Exit(1)
		}
	}

	SetTempFilesPath(options.tempPath, options.unencryptedTempPath) // Set temporal files path

	if options.clean {
		LogInfo("Cleaning temporal files...")
		ClearTemporalFilesPath()
	}

	if options.initialize {
		InitializeCredentialsPath(options.vaultPath)
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

		// Create and initialize vault

		vault := Vault{}
		err := vault.Initialize(options.vaultPath)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		GLOBAL_VAULT = &vault

		// Create and run HTTP server
		RunHTTPServer()
	} else if options.initialize || options.clean {
		return
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
	fmt.Println("        --init, -i                 Initializes the vault. Asks for username and password.")
	fmt.Println("        --debug                    Enables debug mode.")
	fmt.Println("        --cors-insecure            Allows all CORS requests (insecure, for development).")
	fmt.Println("        --vault-path, -vp <path>   Sets the data storage path for the vault.")
	fmt.Println("        --clean, -c                Cleans temporal path before starting the daemon.")
	fmt.Println("    ENVIRONMENT VARIABLES:")
	fmt.Println("        FFMPEG_PATH                Path to ffmpeg binary.")
	fmt.Println("        FFPROBE_PATH               Path to ffprobe binary.")
	fmt.Println("        TEMP_PATH                  Temporal path to modify files before pushing them to the vault.")
	fmt.Println("                                   Note: It should be in the same filesystem as the vault.")
	fmt.Println("        UNENCRYPTED_TEMP_PATH      Temporal path to store things like uploaded files or to use for FFMPEG encoding.")
	fmt.Println("                                   Note: It should be in a different filesystem if the vault is stored in an unsafe environment.")
	fmt.Println("                                   By default, this will be stored in ~/.pmv/temp")
	fmt.Println("        FRONTEND_PATH              Path to static frontend.")
	fmt.Println("        BIND_ADDRESS               Bind address for listening HTTP and HTTPS.")
	fmt.Println("        HTTP_PORT                  HTTP listening port, 80 by default.")
	fmt.Println("        SSL_PORT                   HTTPS listening port, 443 by default.")
	fmt.Println("        SSL_CERT                   HTTPS certificate (.pem) path.")
	fmt.Println("        SSL_KEY                    HTTPS private key (.pem) path.")
	fmt.Println("        USING_PROXY                Set to 'YES' if you are using a reverse proxy.")
}

func printVersion() {
	fmt.Println("PersonalMediaVault 1.0.0")
}
