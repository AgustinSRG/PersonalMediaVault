// Main test (Integration)

package main

import (
	"errors"
	"os"
	"path"
	"runtime"
	"sync"
	"testing"

	child_process_manager "github.com/AgustinSRG/go-child-process-manager"
	"github.com/gorilla/mux"
)

var vaultMutex = &sync.Mutex{}
var testVaultInitialized = false

var apiRouter *mux.Router

func ErrorMismatch(propName string, val string, expectedVal string) error {
	return errors.New("Value mismatch. Expected '" + expectedVal + "' but found '" + val + "' for " + propName)
}

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

func TestAPIIntegration(t *testing.T) {
	server, err := GetTestServer()

	if err != nil {
		t.Error(err)
		return
	}

	defer server.Close()

	// Login

	session, _, err := LoginTest(server)

	if err != nil {
		t.Error(err)
		return
	}

	// API tests

	wg := sync.WaitGroup{}

	wg.Add(1)
	t.Run("Account API", func(t *testing.T) {
		AccountAPITest(server, session, t)
		wg.Done()
	})

	// Wait for API tests to finish before logging out

	wg.Wait()

	// Logout

	err = LogoutTest(server, session)

	if err != nil {
		t.Error(err)
		return
	}
}
