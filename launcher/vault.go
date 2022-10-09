// Vault controller

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
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
	launchTag      string

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

	port := fmt.Sprint(vc.launchConfig.Port)
	bindAddr := "[::]"

	if vc.launchConfig.Local {
		bindAddr = "[::1]"
	}

	fmt.Println("Vault path: " + vc.vaultPath)
	fmt.Println("Vault listening address: " + bindAddr + ":" + port)

	if vc.launchConfig.hasSSL() {
		fmt.Println("SSL: Enabled")
	} else {
		fmt.Println("SSL: Disabled")
	}

	if vc.started {
		fmt.Println("Status: Started")
	} else {
		fmt.Println("Status: Stopped")
	}

	if vc.errorMessage != "" {
		fmt.Println("Error: " + vc.errorMessage)
	}

	if vc.logFilePath != "" {
		fmt.Println("Log file: " + vc.logFilePath)
	}
}

func (vc *VaultController) Start() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if vc.started {
		fmt.Println("Vault is already started. Use 'browser' to open it in the browser.")
		return false
	}

	if CheckVaultLocked(path.Join(vc.vaultPath, "vault.lock")) {
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

	// Tag

	vc.launchTag = fmt.Sprint(time.Now().UnixMilli()) + "-" + fmt.Sprint(os.Getpid())

	// Run backend

	fmt.Println("Starting vault...")

	cmd := exec.Command(BACKEND_BIN, "--daemon", "--clean", "--vault-path", vc.vaultPath, "--port", port, "--bind", bindAddr, "--launch-tag", vc.launchTag)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FFMPEG_PATH="+FFMPEG_BIN, "FFPROBE_PATH="+FFPROBE_BIN, "FRONTEND_PATH="+FRONTEND_PATH)

	if vc.launchConfig.hasSSL() {
		cmd.Env = append(cmd.Env, "SSL_CERT="+vc.launchConfig.SSL_Cert, "SSL_KEY="+vc.launchConfig.SSL_Key)
	}

	if vc.launchConfig.SecureTempDelete {
		cmd.Env = append(cmd.Env, "TEMP_FILE_DELETE_MODE=SECURE")
	}

	cmd.Stderr = logFile
	cmd.Stdout = logFile
	cmd.Stdin = nil

	child_process_manager.ConfigureCommand(cmd)

	err = cmd.Start()

	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	// Add process as a child process
	child_process_manager.AddChildProcess(cmd.Process)

	vc.started = true
	vc.errorMessage = ""
	vc.backendProcess = cmd.Process

	go vc.WaitForProcess()

	return true
}

func (vc *VaultController) WaitForStart() bool {
	done := false
	hasErr := false
	errMsg := ""
	tag := ""

	for !done {
		vc.lock.Lock()

		if !vc.started {
			done = true // This means the process exitted
			hasErr = true
			errMsg = vc.errorMessage
		} else {
			tag = vc.launchTag
		}

		vc.lock.Unlock()

		if !done {
			// Check the port availability
			if vc.launchConfig.hasSSL() {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				}
				client := &http.Client{Transport: tr}
				resp, err := client.Get("https://localhost:" + fmt.Sprint(vc.launchConfig.Port) + "/api/admin/launcher/" + tag)

				if err == nil {
					resp.Body.Close()

					if resp.StatusCode == 200 {
						done = true
					}
				}
			} else {
				resp, err := http.Get("http://localhost:" + fmt.Sprint(vc.launchConfig.Port) + "/api/admin/launcher/" + tag)

				if err == nil {
					resp.Body.Close()

					if resp.StatusCode == 200 {
						done = true
					}
				}
			}
		}

		if !done {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if hasErr {
		if errMsg != "" {
			fmt.Println("Error: " + errMsg)
		} else {
			fmt.Println("Could not start the vault. Check logs for details")
		}
	} else {
		fmt.Println("Vault successfully started")
	}

	return !hasErr
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
		} else if state.ExitCode() == 6 {
			vc.errorMessage = "Invalid SSL keypair provided. Please fix it with 'ssl-setup'."
		} else {
			vc.errorMessage = ""
		}
	}

	// Remove lock

	os.Remove(path.Join(vc.vaultPath, "vault.lock"))

	vc.logFile.Close()
	vc.logFile = nil
	vc.logFilePath = ""

	vc.lock.Unlock()
}

func (vc *VaultController) Clean() {
	// Clean
	cmd := exec.Command(BACKEND_BIN, "--clean", "--fix-consistency", "--skip-lock", "--vault-path", vc.vaultPath)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func (vc *VaultController) Backup(p string) {
	fmt.Println("You are going to create a backup of the vault in the following location:")
	fmt.Println(p)
	fmt.Print("Procceed? (y/n): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		return
	}

	// Backup
	cmd := exec.Command(BACKUP_BIN, vc.vaultPath, p)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()

	if err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func (vc *VaultController) SetupSSL() bool {
	fmt.Print("Do you want to setup SSL for your vault? (y/n): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		return false
	}

	fmt.Println("Type the absolute path in your file system to the SSL CERTIFICATE file you want to use.")
	fmt.Println("Make sure it is correct and it is in PEM format.")
	fmt.Print("Certificate file: ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	certFile := ans

	fmt.Println("Type the absolute path in your file system to the SSL PRIVATE KEY file you want to use.")
	fmt.Println("Make sure it is correct and it is in PEM format.")
	fmt.Print("Private key file: ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	keyFile := ans

	fmt.Println("")
	if certFile == "" {
		fmt.Println("SSL certificate file: " + "(Not Set)")
	} else {
		fmt.Println("SSL certificate file: " + certFile)
	}
	if keyFile == "" {
		fmt.Println("SSL key file: " + "(Not Set)")
	} else {
		fmt.Println("SSL key file: " + keyFile)
	}

	fmt.Print("Is this correct? (y/n): ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		return false
	}

	vc.launchConfig.SSL_Cert = certFile
	vc.launchConfig.SSL_Key = keyFile

	if vc.launchConfig.SSL_Cert != "" && vc.launchConfig.SSL_Key != "" {
		if vc.launchConfig.Port == 80 {
			vc.launchConfig.Port = 443
		}
	} else {
		if vc.launchConfig.Port == 443 {
			vc.launchConfig.Port = 80
		}
	}

	err = writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	} else {
		fmt.Println("Changes in configuration successfully saved.")
		return true
	}
}

func (vc *VaultController) disableSSL() bool {
	if !vc.launchConfig.hasSSL() {
		fmt.Println("SSL is not enabled for this vault.")
		return false
	}

	fmt.Print("Do you want to disable SSL for your vault? (y/n): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error: " + err.Error())
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		return false
	}

	vc.launchConfig.SSL_Cert = ""
	vc.launchConfig.SSL_Key = ""

	if vc.launchConfig.Port == 443 {
		vc.launchConfig.Port = 80
	}

	err = writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	} else {
		fmt.Println("Changes in configuration successfully saved.")
		return true
	}
}

func (vc *VaultController) SetSecureTempDelete(d bool) bool {
	vc.launchConfig.SecureTempDelete = d

	if d {
		fmt.Println("Secure deletion of temp files is now ENABLED.")
	} else {
		fmt.Println("Secure deletion of temp files is now DISABLED.")
	}

	err := writeLauncherConfig(path.Join(vc.vaultPath, "launcher.config.json"), vc.launchConfig)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		return false
	} else {
		fmt.Println("Changes in configuration successfully saved.")
		return true
	}
}
