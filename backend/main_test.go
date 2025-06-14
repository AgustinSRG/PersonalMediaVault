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
var testVaultKey []byte

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
			ffmpegPath = "C:\\\\ffmpeg\\bin\\ffmpeg.exe"
		} else {
			ffmpegPath = "/usr/bin/ffmpeg"
		}
	}

	ffprobePath := os.Getenv("FFPROBE_PATH")

	if ffprobePath == "" {
		if runtime.GOOS == "windows" {
			ffprobePath = "C:\\\\ffmpeg\\bin\\ffprobe.exe"
		} else {
			ffprobePath = "/usr/bin/ffprobe"
		}
	}

	SetFFMPEGBinaries(ffmpegPath, ffprobePath) // Set FFMPEG paths

	if os.Getenv("LOG_DEBUG") == "YES" {
		SetDebugLogEnabled(true) // Log debug mode
	}

	if os.Getenv("LOG_REQUESTS") == "YES" {
		SetRequestLogEnabled(true) // Log requests
	}

	SetTempFilesPath(path.Join("test-vault", "temp"))

	ClearTemporalFilesPath()

	vault := Vault{}
	err = vault.Initialize("test-vault", 2)

	if err != nil {
		return err
	}

	testVaultInitialized = true

	GLOBAL_VAULT = &vault

	SetUnencryptedTempFilesPath("temp")

	ClearUnencryptedTempFilesPath()

	apiRouter = RunHTTPServer("", "", true)

	initialUser := os.Getenv("VAULT_INITIAL_USER")

	if initialUser == "" {
		initialUser = VAULT_DEFAULT_USER
	}

	initialPassword := os.Getenv("VAULT_INITIAL_PASSWORD")

	if initialPassword == "" {
		initialPassword = VAULT_DEFAULT_PASSWORD
	}

	key, _, _, err := vault.credentials.UnlockVault(initialUser, initialPassword)

	if err != nil {
		return err
	}

	testVaultKey = key

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
		Account_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Admin API", func(t *testing.T) {
		Admin_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Media API", func(t *testing.T) {
		Media_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Tags API", func(t *testing.T) {
		Tags_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Search API", func(t *testing.T) {
		Search_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Albums API", func(t *testing.T) {
		Albums_API_Test(server, session, t)
		wg.Done()
	})

	wg.Add(1)
	t.Run("Config API", func(t *testing.T) {
		Config_API_Test(server, session, t)
		wg.Done()
	})

	// Wait for API tests to finish before logging out

	wg.Wait()

	// Wait for tasks to finish

	wg2 := sync.WaitGroup{}

	wg2.Add(1)
	t.Run("Tasks API", func(t *testing.T) {
		Tasks_API_Test(server, session, t)
		wg2.Done()
	})

	wg2.Wait()

	// Logout

	err = LogoutTest(server, session)

	if err != nil {
		t.Error(err)
		return
	}
}
