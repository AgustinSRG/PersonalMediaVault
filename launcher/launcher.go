// Launcher tools

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var (
	BACKEND_BIN   = ""
	BACKUP_BIN    = ""
	FFMPEG_BIN    = ""
	FFPROBE_BIN   = ""
	FRONTEND_PATH = ""
)

type LauncherConfig struct {
	Port     int    `json:"port"`
	Local    bool   `json:"local"`
	SSL_Cert string `json:"ssl_cert"`
	SSL_Key  string `json:"ssl_key"`
}

func (c *LauncherConfig) hasSSL() bool {
	return c.SSL_Cert != "" && c.SSL_Key != ""
}

func detectLauncherPaths() {

	// Backend

	BACKEND_BIN = path.Join(getDirName(), "bin", getBinaryFileName("pmvd"))

	if !fileExists(BACKEND_BIN) {
		BACKEND_BIN = path.Join("/usr/bin", getBinaryFileName("pmvd"))

		if !fileExists(BACKEND_BIN) {
			BACKEND_BIN = path.Join("../backend", getBinaryFileName("pmvd"))

			if !fileExists(BACKEND_BIN) {
				fmt.Println("Error: Could not find the backend binary (pmvd)")
				fmt.Println("Seems like some required files are missing.")
				fmt.Println("In order to fix this error, you could re-install PersonalMediaVault.")
				os.Exit(1)
			}
		}
	}

	// Backup tool

	BACKUP_BIN = path.Join(getDirName(), "bin", getBinaryFileName("pmv-backup"))

	if !fileExists(BACKUP_BIN) {
		BACKUP_BIN = path.Join("/usr/bin", getBinaryFileName("pmv-backup"))

		if !fileExists(BACKUP_BIN) {
			BACKUP_BIN = path.Join("../backup-tool", getBinaryFileName("pmv-backup"))

			if !fileExists(BACKUP_BIN) {
				fmt.Println("Error: Could not find the backup tool binary (pmv-backup)")
				fmt.Println("Seems like some required files are missing.")
				fmt.Println("In order to fix this error, you could re-install PersonalMediaVault.")
				os.Exit(1)
			}
		}
	}

	// FFMPEG

	FFMPEG_BIN = path.Join(getDirName(), "bin", getBinaryFileName("ffmpeg"))

	if !fileExists(FFMPEG_BIN) {
		FFMPEG_BIN = path.Join("/usr/bin", getBinaryFileName("ffmpeg"))

		if !fileExists(FFMPEG_BIN) {
			FFMPEG_BIN = path.Join("/ffmpeg/bin/", getBinaryFileName("ffmpeg"))

			if !fileExists(FFMPEG_BIN) {
				fmt.Println("Error: Could not find the ffmpeg binary (ffmpeg)")
				fmt.Println("Seems like some required files are missing.")
				fmt.Println("In order to fix this error, you could re-install PersonalMediaVault.")
				os.Exit(1)
			}
		}
	}

	// FFPROBE

	FFPROBE_BIN = path.Join(getDirName(), "bin", getBinaryFileName("ffprobe"))

	if !fileExists(FFPROBE_BIN) {
		FFPROBE_BIN = path.Join("/usr/bin", getBinaryFileName("ffprobe"))

		if !fileExists(FFPROBE_BIN) {
			FFPROBE_BIN = path.Join("/ffmpeg/bin/", getBinaryFileName("ffprobe"))

			if !fileExists(FFPROBE_BIN) {
				fmt.Println("Error: Could not find the ffprobe binary (ffprobe)")
				fmt.Println("Seems like some required files are missing.")
				fmt.Println("In order to fix this error, you could re-install PersonalMediaVault.")
				os.Exit(1)
			}
		}
	}

	// Front

	FRONTEND_PATH = path.Join(getDirName(), "www")
	if !folderExists(FRONTEND_PATH) {
		FRONTEND_PATH = "/usr/lib/pmv/www"
		if !folderExists(FRONTEND_PATH) {
			FRONTEND_PATH = "../frontend/dist"

			if !folderExists(FRONTEND_PATH) {
				fmt.Println("Error: Could not find the frontend package directory")
				fmt.Println("Seems like some required files are missing.")
				fmt.Println("In order to fix this error, you could re-install PersonalMediaVault.")
				os.Exit(1)
			}
		}
	}
}

func readLauncherConfig(file string) LauncherConfig {
	b, err := ioutil.ReadFile(file)

	if err != nil {
		return LauncherConfig{
			Port:  0,
			Local: true,
		}
	}

	config := LauncherConfig{
		Port:  0,
		Local: true,
	}

	// Parse
	json.Unmarshal(b, &config)

	return config
}

func writeLauncherConfig(file string, config LauncherConfig) error {
	// Get the json data
	jsonData, err := json.Marshal(config)

	if err != nil {
		return err
	}

	// Write file
	err = ioutil.WriteFile(file, jsonData, FILE_PERMISSION)
	if err != nil {
		return err
	}

	return nil
}
