// Vault controller

package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type VaultController struct {
	vaultPath     string
	launchConfig  LauncherConfig
	consoleReader *bufio.Reader

	started        bool
	errorMessage   string
	backendProcess *os.Process
	logFilePath    string
	logFile        *os.File

	lock *sync.Mutex
}

func (vc *VaultController) Initialize(vaultPath string, launchConfig LauncherConfig, consoleReader *bufio.Reader) {
	vc.vaultPath = vaultPath
	vc.launchConfig = launchConfig
	vc.consoleReader = consoleReader

	vc.lock = &sync.Mutex{}

	vc.started = false
	vc.errorMessage = ""
	vc.backendProcess = nil
	vc.logFile = nil
	vc.logFilePath = ""

	if !fileExists(path.Join(vaultPath, "credentials.json")) {
		// Run initialize sequence
		cmd := exec.Command(BACKEND_BIN, "--init", "--skip-lock", "--vault-path", vaultPath)

		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		err := cmd.Run()

		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}
	}
}

func (vc *VaultController) PrintStatus() {
	vc.lock.Lock()
	defer vc.lock.Unlock()
}

func (vc *VaultController) Start() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if vc.started {
		fmt.Println("Vault is already started. Use 'browser' to open it in the browser.")
		return false
	}

	if fileExists(path.Join(vc.vaultPath, "vault.lock")) {
		fmt.Println("Seems like the vault is being used by another process.")
		fmt.Println("Openning the vault by multiple processes could be dangerous for the vault integrity.")
		fmt.Print("Procceed? (y/n): ")

		ans, err := vc.consoleReader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: " + err.Error())
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if strings.HasPrefix(strings.ToLower(ans), "y") {
			os.Remove(path.Join(vc.vaultPath, "vault.lock"))
		} else {
			return false
		}
	}

	// Make logfile

	logFilePath, err := getLogFileName()

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	vc.logFilePath = logFilePath

	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	vc.logFile = logFile

	// Options

	port := fmt.Sprint(vc.launchConfig.Port)
	bindAddr := "[::]"

	if vc.launchConfig.Local {
		bindAddr = "[::1]"
	}

	// Run backend

	fmt.Println("Starting vault...")

	cmd := exec.Command(BACKEND_BIN, "--daemon", "--clean", "--vault-path", vc.vaultPath, "--port", port, "--bind", bindAddr)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FFMPEG_PATH="+FFMPEG_BIN, "FFPROBE_PATH="+FFPROBE_BIN, "FRONTEND_PATH="+FRONTEND_PATH)

	cmd.Stderr = logFile
	cmd.Stdout = logFile
	cmd.Stdin = nil

	err = cmd.Start()

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	vc.started = true
	vc.backendProcess = cmd.Process

	go vc.WaitForProcess()

	return true
}

func (vc *VaultController) WaitForStart() {
	done := false

	for !done {
		vc.lock.Lock()

		if !vc.started {
			done = true // This means the process exitted
		}

		vc.lock.Unlock()

		if !done {
			// Check the port availability
			resp, err := http.Get("http://localhost:" + fmt.Sprint(vc.launchConfig.Port) + "/")

			if err == nil {
				resp.Body.Close()

				if resp.StatusCode == 200 {
					done = true
				}
			}
		}

		if !done {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("Vault successfully started")
}

func (vc *VaultController) Stop() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if !vc.started || vc.backendProcess == nil {
		fmt.Println("Vault is already stopped.")
		return false
	}

	fmt.Println("Stopping vault...")

	vc.backendProcess.Kill()

	return true
}

func (vc *VaultController) WaitForStop() {
	done := false

	for !done {
		vc.lock.Lock()

		if !vc.started {
			done = true
		}

		vc.lock.Unlock()

		if !done {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("Vault successfully stopped")
}

func (vc *VaultController) WaitForProcess() {
	vc.lock.Lock()
	p := vc.backendProcess
	vc.lock.Unlock()

	if p == nil {
		return
	}

	state, err := p.Wait()

	vc.lock.Lock()

	vc.started = false
	vc.backendProcess = nil

	if err != nil {
		vc.errorMessage = err.Error()
	} else {
		if state.ExitCode() == 0 {
			vc.errorMessage = ""
		} else if state.ExitCode() == 4 {
			vc.errorMessage = "Vault is locked by another process. Cannot start."
		} else if state.ExitCode() == 5 {
			vc.errorMessage = "Cannot listen to port " + fmt.Sprint(vc.launchConfig.Port) + ". Probably there is another proccess listening on that port."
		} else {
			vc.errorMessage = "Backend process error. Check the logs at: " + vc.logFilePath
		}
	}

	// Remove lock

	os.Remove(path.Join(vc.vaultPath, "vault.lock"))

	vc.logFile.Close()
	vc.logFile = nil
	vc.logFilePath = ""

	vc.lock.Unlock()
}
