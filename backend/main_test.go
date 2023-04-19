// Main test utils

package main

import (
	"os"
	"path"
	"runtime"
	"sync"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/gorilla/mux"
)

var vaultMutex = &sync.Mutex{}
var testVaultInitialized = false

var apiRouter *mux.Router

func InitializeTestVault() error {
	vaultMutex.Lock()
	defer vaultMutex.Unlock()

	if testVaultInitialized {
		return nil
	}

	os.RemoveAll("test-vault") // Remove test vault before starting

	err := child_process_manager.InitializeChildProcessManager()

	if err != nil {
		return err
	}

	ffmpegPath := os.Getenv("FFMPEG_PATH")
	if ffmpegPath == "" {
		if runtime.GOOS == "windows" {
			ffmpegPath = "/ffmpeg/bin/ffmpeg.exe"
		} else {
			ffmpegPath = "/usr/bin/ffmpeg"
		}
	}

	ffprobePath := os.Getenv("FFPROBE_PATH")

	if ffprobePath == "" {
		if runtime.GOOS == "windows" {
			ffprobePath = "/ffmpeg/bin/ffprobe.exe"
		} else {
			ffprobePath = "/usr/bin/ffprobe"
		}
	}

	SetFFMPEGBinaries(ffmpegPath, ffprobePath) // Set FFMPEG paths

	SetDebugLogEnabled(true)   // Log debug mode
	SetRequestLogEnabled(true) // Log requests

	SetTempFilesPath(path.Join("test-vault", "temp"))

	ClearTemporalFilesPath()

	vault := Vault{}
	err = vault.Initialize("test-vault")

	if err != nil {
		return err
	}

	testVaultInitialized = true

	GLOBAL_VAULT = &vault

	SetUnencryptedTempFilesPath("temp")

	apiRouter = RunHTTPServer("", "", true)

	return nil
}
