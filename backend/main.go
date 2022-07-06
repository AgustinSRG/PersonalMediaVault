// Main

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const BACKEND_VERSION = "1.0.0"

type BackendOptions struct {
	debug       bool // Debug mode
	logRequests bool

	// Run modes
	daemon     bool
	initialize bool
	clean      bool

	// Lock
	skipLock bool

	// FFmpeg
	ffmpegPath  string
	ffprobePath string

	// Vault
	vaultPath string

	// Temp path
	tempPath            string
	unencryptedTempPath string

	// Open browser
	openBrowser int
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
		tempPath:            "./vault/temp",
		unencryptedTempPath: os.Getenv("TEMP_PATH"),
		openBrowser:         0,
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
		} else if arg == "--log-requests" {
			options.logRequests = true
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
		} else if arg == "--skip-lock" || arg == "-sl" {
			options.skipLock = true
		} else if arg == "--vault-path" || arg == "-vp" {
			if i == len(args)-1 {
				fmt.Println("The option '--vault-path' requires a value")
				os.Exit(1)
			}
			options.vaultPath = args[i+1]
			options.tempPath = path.Join(options.vaultPath, "temp")
			i++
		} else if arg == "--open-browser" {
			options.openBrowser = 1
		} else if arg == "--open-browser-ssl" {
			options.openBrowser = 2
		} else {
			fmt.Println("Invalid argument: " + arg)
			os.Exit(1)
		}
	}

	if options.daemon || options.clean || options.initialize {
		printVersion()
		if !options.skipLock {
			// Setup lockfile
			if !TryLockVault(options.vaultPath) {
				fmt.Println("Error: The vault is already in use. If this is not true, please remove the file vault.lock in your vault folder.")
				os.Exit(1)
			}
		}
	}

	if options.initialize {
		InitializeCredentialsPath(options.vaultPath)
	}

	SetTempFilesPath(options.tempPath) // Set temporal files path

	if options.clean {
		LogInfo("Cleaning vault temporal files...")
		ClearTemporalFilesPath()
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

		SetDebugLogEnabled(options.debug)         // Log debug mode
		SetRequestLogEnabled(options.logRequests) // Log requests

		// Create and initialize vault

		vault := Vault{}
		err := vault.Initialize(options.vaultPath)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		absolutePath, err := filepath.Abs(options.vaultPath)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		LogInfo("Openned vault: " + absolutePath)

		GLOBAL_VAULT = &vault

		options.unencryptedTempPath = path.Join(options.unencryptedTempPath, vault.credentials.GetFingerprint())

		SetUnencryptedTempFilesPath(options.unencryptedTempPath)

		if options.clean {
			LogInfo("Cleaning unencrypted temporal files...")
			ClearUnencryptedTempFilesPath()
		}

		// Create and run HTTP server
		if options.openBrowser > 0 {
			go openBrowser(options.openBrowser)
		}
		RunHTTPServer()
	} else if options.initialize || options.clean {
		vault := Vault{}
		err := vault.Initialize(options.vaultPath)

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		options.unencryptedTempPath = path.Join(options.unencryptedTempPath, vault.credentials.GetFingerprint())

		SetUnencryptedTempFilesPath(options.unencryptedTempPath)

		if options.clean {
			LogInfo("Cleaning unencrypted temporal files...")
			ClearUnencryptedTempFilesPath()
		}

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
	fmt.Println("        --log-requests             Enables logging requests to standard outout.")
	fmt.Println("        --cors-insecure            Allows all CORS requests (insecure, for development).")
	fmt.Println("        --vault-path, -vp <path>   Sets the data storage path for the vault.")
	fmt.Println("        --clean, -c                Cleans temporal path before starting the daemon.")
	fmt.Println("        --skip-lock, -sl           Ignores vault lockfile.")
	fmt.Println("        --open-browser             Opens browser in localhost (HTTP).")
	fmt.Println("        --open-browser-ssl         Opens browser in localhost (HTTPS).")
	fmt.Println("    ENVIRONMENT VARIABLES:")
	fmt.Println("        FFMPEG_PATH                Path to ffmpeg binary.")
	fmt.Println("        FFPROBE_PATH               Path to ffprobe binary.")
	fmt.Println("        TEMP_PATH                  Temporal path to store things like uploaded files or to use for FFMPEG encoding.")
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
	fmt.Println("---------------------------------------------------")
	fmt.Println("-  _____    __  __  __      __")
	fmt.Println("- |  __ \\  |  \\/  | \\ \\    / /")
	fmt.Println("- | |__) | | \\  / |  \\ \\  / /")
	fmt.Println("- |  ___/  | |\\/| |   \\ \\/ /")
	fmt.Println("- | |      | |  | |    \\  /")
	fmt.Println("- |_|      |_|  |_|     \\/")
	fmt.Println("---------------------------------------------------")
	fmt.Println("- Personal Media Vault")
	fmt.Println("- Version " + BACKEND_VERSION)
	fmt.Println("- https://github.com/AgustinSRG/PersonalMediaVault")
	fmt.Println("---------------------------------------------------")
}

func openBrowser(opt int) {
	// Generate localhost URL
	var url string

	if opt == 2 {
		var ssl_port int
		ssl_port = 443
		customSSLPort := os.Getenv("SSL_PORT")
		if customSSLPort != "" {
			sslp, e := strconv.Atoi(customSSLPort)
			if e == nil {
				ssl_port = sslp
			}
		}

		if ssl_port == 443 {
			url = "https://localhost"
		} else {
			url = "https://localhost:" + fmt.Sprint(ssl_port)
		}
	} else {
		var tcp_port int
		tcp_port = 80
		customTCPPort := os.Getenv("HTTP_PORT")
		if customTCPPort != "" {
			tcpp, e := strconv.Atoi(customTCPPort)
			if e == nil {
				tcp_port = tcpp
			}
		}

		if tcp_port == 80 {
			url = "http://localhost"
		} else {
			url = "http://localhost:" + fmt.Sprint(tcp_port)
		}
	}

	// Wait a bit so the server can start (1 second)
	time.Sleep(1 * time.Second)

	LogInfo("Openning frontend URL: " + url)

	// Open the browser
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		LogWarning("Error open browser: " + err.Error())
	}
}
