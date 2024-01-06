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
	"syscall"
	"time"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/term"
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
		// Ask for initial credentials
		reader := bufio.NewReader(os.Stdin)

		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNotFoundCreateOne",
				Other: "Vault does not exists. Please provide a set of credentials to create one.",
			},
		})

		fmt.Println(msg)

		var username string = ""

		for username == "" {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "EnterUsername",
					Other: "Enter Username",
				},
			})
			fmt.Print(msg + ": ")
			readUsername, err := reader.ReadString('\n')
			if err != nil {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Error",
						Other: "Error: {{.Message}}",
					},
					TemplateData: map[string]interface{}{
						"Message": err.Error(),
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}

			username = strings.TrimSpace(readUsername)

			if username == "" {
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorUsernameBlank",
						Other: "Username cannot be blank.",
					},
				})
				fmt.Println(msg)
				continue
			}

			if len(username) > 255 {
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorUsernameTooLong",
						Other: "Username cannot be longer than 255 characters.",
					},
				})
				fmt.Println(msg)
				username = ""
				continue
			}
		}

		var password string = ""
		var password_repeat string = ""

		for password == "" || password != password_repeat {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "EnterPassword",
					Other: "Enter Password",
				},
			})
			fmt.Print(msg + ": ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Error",
						Other: "Error: {{.Message}}",
					},
					TemplateData: map[string]interface{}{
						"Message": err.Error(),
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}

			password = strings.TrimSpace(string(bytePassword))

			if password == "" {
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorPasswordBlank",
						Other: "Password cannot be blank.",
					},
				})
				fmt.Println(msg)
				continue
			}

			if len(password) > 255 {
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorPasswordTooLong",
						Other: "Password cannot be longer than 255 characters.",
					},
				})
				fmt.Println(msg)
				password = ""
				continue
			}

			fmt.Print("\n")

			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "RepeatPassword",
					Other: "Repeat Password",
				},
			})
			fmt.Print(msg + ": ")
			bytePassword, err = term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "Error",
						Other: "Error: {{.Message}}",
					},
					TemplateData: map[string]interface{}{
						"Message": err.Error(),
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}

			fmt.Print("\n")

			password_repeat = strings.TrimSpace(string(bytePassword))

			if password != password_repeat {
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorPasswordNotMatching",
						Other: "Passwords do not match.",
					},
				})
				fmt.Println(msg)
			}
		}

		// Run initialize sequence
		cmd := exec.Command(BACKEND_BIN, "--init", "--skip-lock", "--vault-path", vaultPath)

		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "PMV_INIT_SET_USER="+username, "PMV_INIT_SET_PASSWORD="+password)

		cmd.Stderr = os.Stderr
		cmd.Stdout = nil
		cmd.Stdin = nil

		err := cmd.Run()

		if err != nil {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Error",
					Other: "Error: {{.Message}}",
				},
				TemplateData: map[string]interface{}{
					"Message": err.Error(),
				},
			})
			fmt.Println(msg)
			os.Exit(1)
		}

		if cmd.ProcessState.ExitCode() == 0 {
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "VaultInitialized",
					Other: "Vault initialized successfully!",
				},
			})
			fmt.Println(msg)
		} else {
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
		bindAddr = "localhost"
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultPath",
			Other: "Vault path: {{.Path}}",
		},
		TemplateData: map[string]interface{}{
			"Path": vc.vaultPath,
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultListeningAddress",
			Other: "Vault listening address: {{.Address}}",
		},
		TemplateData: map[string]interface{}{
			"Address": bindAddr + ":" + port,
		},
	})
	fmt.Println(msg)

	if vc.launchConfig.hasSSL() {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLEnabled",
				Other: "SSL: Enabled",
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLDisabled",
				Other: "SSL: Disabled",
			},
		})
		fmt.Println(msg)
	}

	if vc.started {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "StatusStarted",
				Other: "Status: Started",
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "StatusStopped",
				Other: "Status: Stopped",
			},
		})
		fmt.Println(msg)
	}

	if vc.errorMessage != "" {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": vc.errorMessage,
			},
		})
		fmt.Println(msg)
	}

	if vc.logFilePath != "" {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LogFile",
				Other: "Log file: {{.File}}",
			},
			TemplateData: map[string]interface{}{
				"File": vc.logFilePath,
			},
		})
		fmt.Println(msg)
	}
}

func (vc *VaultController) Start() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if vc.started {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultAlreadyStarted",
				Other: "Vault is already started. Use 'browser' to open it in the browser.",
			},
		})
		fmt.Println(msg)
		return false
	}

	if CheckVaultLocked(path.Join(vc.vaultPath, "vault.lock")) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultBeingUsed",
				Other: "Seems like the vault is being used by another process",
			},
		})
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "OpenMultipleRisk",
				Other: "Opening the vault by multiple processes could be dangerous for the vault integrity.",
			},
		})
		fmt.Println(msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Proceed",
				Other: "Proceed?",
			},
		})
		ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YesNo",
				Other: "y/n",
			},
		})
		fmt.Print(msg + " (" + ynMsg + "): ")

		ans, err := vc.consoleReader.ReadString('\n')
		if err != nil {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Error",
					Other: "Error: {{.Message}}",
				},
				TemplateData: map[string]interface{}{
					"Message": err.Error(),
				},
			})
			fmt.Println(msg)
			os.Exit(1)
		}

		ans = strings.TrimSpace(ans)

		if checkYesNoAnswer(ans) {
			os.Remove(path.Join(vc.vaultPath, "vault.lock"))
		} else {
			return false
		}
	}

	// Make log file

	logFilePath, err := getLogFileName()

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	vc.logFilePath = logFilePath

	logFile, err := os.Create(logFilePath)
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	vc.logFile = logFile

	// Options

	port := fmt.Sprint(vc.launchConfig.Port)
	bindAddr := ""

	if vc.launchConfig.Local {
		bindAddr = "127.0.0.1"
	}

	// Tag

	vc.launchTag = fmt.Sprint(time.Now().UnixMilli()) + "-" + fmt.Sprint(os.Getpid())

	// Cache size

	cacheSize := DEFAULT_CACHE_SIZE

	if vc.launchConfig.CacheSize != nil {
		cacheSize = *(vc.launchConfig.CacheSize)
	}

	if cacheSize < 0 {
		cacheSize = 0
	}

	// Run backend

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "StartingVault",
			Other: "Starting vault...",
		},
	})
	fmt.Println(msg)

	cmd := exec.Command(BACKEND_BIN, "--daemon", "--clean", "--vault-path", vc.vaultPath, "--port", port, "--bind", bindAddr, "--launch-tag", vc.launchTag, "--cache-size", fmt.Sprint(cacheSize))

	if vc.launchConfig.LogRequests {
		cmd.Args = append(cmd.Args, "--log-requests")
	}

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

	err = child_process_manager.ConfigureCommand(cmd)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	err = cmd.Start()

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	// Add process as a child process
	err = child_process_manager.AddChildProcess(cmd.Process)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

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
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "Error",
					Other: "Error: {{.Message}}",
				},
				TemplateData: map[string]interface{}{
					"Message": errMsg,
				},
			})
			fmt.Println(msg)
		} else {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "VaultStartError",
					Other: "Could not start the vault. Check logs for details",
				},
			})
			fmt.Println(msg)
		}
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultStartedSuccessfully",
				Other: "Vault successfully started.",
			},
		})
		fmt.Println(msg)
	}

	return !hasErr
}

func (vc *VaultController) Stop() bool {
	vc.lock.Lock()
	defer vc.lock.Unlock()

	if !vc.started || vc.backendProcess == nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultAlreadyStopped",
				Other: "Vault is already stopped.",
			},
		})
		fmt.Println(msg)
		return false
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultStopping",
			Other: "Stopping vault...",
		},
	})
	fmt.Println(msg)

	vc.backendProcess.Kill() //nolint:errcheck

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

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultStoppedSuccessfully",
			Other: "Vault successfully stopped.",
		},
	})
	fmt.Println(msg)
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
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorVaultLocked",
					Other: "Vault is locked by another process. Cannot start.",
				},
			})
			vc.errorMessage = msg
		} else if state.ExitCode() == 5 {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorPortBusy",
					Other: "Cannot listen to port {{.Port}}. Probably there is another process listening on that port.",
				},
				TemplateData: map[string]interface{}{
					"Port": fmt.Sprint(vc.launchConfig.Port),
				},
			})
			vc.errorMessage = msg
		} else if state.ExitCode() == 6 {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorInvalidKeyPair",
					Other: "Invalid SSL key pair provided. Please fix it with 'ssl-setup'.",
				},
			})
			vc.errorMessage = msg
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
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
	}
}

func (vc *VaultController) RecoverAssets() {
	// Recover
	cmd := exec.Command(BACKEND_BIN, "--recover", "--skip-lock", "--vault-path", vc.vaultPath)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
	}
}

func (vc *VaultController) Backup(p string, re_encrypt bool) {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "AboutToCreateBackup",
			Other: "You are going to create a backup of the vault in the following location:",
		},
	})
	fmt.Println(msg)
	fmt.Println(p)
	if re_encrypt {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "BackupReEncryptNotice",
				Other: "Warning: You are using re-encryption mode. This will change the vault encryption key. Also, any additional accounts won't be moved to the backup.",
			},
		})
		fmt.Println(msg)
	}
	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Proceed",
			Other: "Proceed?",
		},
	})
	ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !checkYesNoAnswer(ans) {
		return
	}

	// Backup
	var cmd *exec.Cmd

	if re_encrypt {
		cmd = exec.Command(BACKUP_BIN, vc.vaultPath, p, "--re-encrypt")
	} else {
		cmd = exec.Command(BACKUP_BIN, vc.vaultPath, p)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PMV_LANGUAGE="+Language)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
	}
}

func (vc *VaultController) SetupSSL() bool {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "SetupSSLAsk",
			Other: "Do you want to setup SSL for your vault?",
		},
	})
	ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !checkYesNoAnswer(ans) {
		return false
	}

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "TypeCertificatePath",
			Other: "Type the absolute path in your file system to the SSL CERTIFICATE file you want to use.",
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "EnsurePEM",
			Other: "Make sure it is correct and it is in PEM format.",
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "CertificateFile",
			Other: "Certificate file",
		},
	})
	fmt.Print(msg + ": ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	certFile := ans

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "TypePrivateKeyPath",
			Other: "Type the absolute path in your file system to the SSL PRIVATE KEY file you want to use",
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "EnsurePEM",
			Other: "Make sure it is correct and it is in PEM format.",
		},
	})
	fmt.Println(msg)

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "PrivateKeyFile",
			Other: "Private key file",
		},
	})
	fmt.Print(msg + ": ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	keyFile := ans

	fmt.Println("")
	if certFile == "" {
		notSetMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NotSet",
				Other: "(Not Set)",
			},
		})
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLCertFile",
				Other: "SSL certificate file: {{.File}}",
			},
			TemplateData: map[string]interface{}{
				"File": notSetMsg,
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLCertFile",
				Other: "SSL certificate file: {{.File}}",
			},
			TemplateData: map[string]interface{}{
				"File": certFile,
			},
		})
		fmt.Println(msg)
	}
	if keyFile == "" {
		notSetMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "NotSet",
				Other: "(Not Set)",
			},
		})
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLkeyFile",
				Other: "SSL key file: {{.File}}",
			},
			TemplateData: map[string]interface{}{
				"File": notSetMsg,
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLkeyFile",
				Other: "SSL key file: {{.File}}",
			},
			TemplateData: map[string]interface{}{
				"File": keyFile,
			},
		})
		fmt.Println(msg)
	}

	msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "IsCorrectAsk",
			Other: "Is this correct?",
		},
	})
	ynMsg, _ = Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err = vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !checkYesNoAnswer(ans) {
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

	err = writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		return false
	} else {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfigChangesSaved",
				Other: "Changes in configuration successfully saved.",
			},
		})
		fmt.Println(msg)
		return true
	}
}

func (vc *VaultController) disableSSL() bool {
	if !vc.launchConfig.hasSSL() {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SSLNotEnabled",
				Other: "SSL is not enabled for this vault.",
			},
		})
		fmt.Println(msg)
		return false
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "DisableSSLAsk",
			Other: "Do you want to disable SSL for your vault?",
		},
	})
	ynMsg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YesNo",
			Other: "y/n",
		},
	})
	fmt.Print(msg + " (" + ynMsg + "): ")

	ans, err := vc.consoleReader.ReadString('\n')
	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		os.Exit(1)
	}

	ans = strings.TrimSpace(ans)

	if !checkYesNoAnswer(ans) {
		return false
	}

	vc.launchConfig.SSL_Cert = ""
	vc.launchConfig.SSL_Key = ""

	if vc.launchConfig.Port == 443 {
		vc.launchConfig.Port = 80
	}

	err = writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		return false
	} else {
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfigChangesSaved",
				Other: "Changes in configuration successfully saved.",
			},
		})
		fmt.Println(msg)
		return true
	}
}

func (vc *VaultController) SetSecureTempDelete(d bool) bool {
	vc.launchConfig.SecureTempDelete = d

	if d {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SecureDeleteEnabled",
				Other: "Secure deletion of temp files is now ENABLED.",
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SecureDeleteDisabled",
				Other: "Secure deletion of temp files is now DISABLED.",
			},
		})
		fmt.Println(msg)
	}

	err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		return false
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfigChangesSaved",
				Other: "Changes in configuration successfully saved.",
			},
		})
		fmt.Println(msg)
		return true
	}
}

func (vc *VaultController) SetLogRequests(d bool) bool {
	vc.launchConfig.LogRequests = d

	if d {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LogRequestsEnabled",
				Other: "Requests logging is now ENABLED.",
			},
		})
		fmt.Println(msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "LogRequestsDisabled",
				Other: "Requests logging is now DISABLED.",
			},
		})
		fmt.Println(msg)
	}

	err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		return false
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfigChangesSaved",
				Other: "Changes in configuration successfully saved.",
			},
		})
		fmt.Println(msg)
		return true
	}
}

func (vc *VaultController) SetCacheSize(s int) bool {
	vc.launchConfig.CacheSize = &s

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "CacheSizeSet",
			Other: "Cache size set to {{.Elements}} elements.",
		},
		TemplateData: map[string]interface{}{
			"Elements": fmt.Sprint(s),
		},
	})
	fmt.Println(msg)

	err := writeLauncherConfig(getLauncherConfigFile(vc.vaultPath), vc.launchConfig)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Println(msg)
		return false
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ConfigChangesSaved",
				Other: "Changes in configuration successfully saved.",
			},
		})
		fmt.Println(msg)
		return true
	}
}
