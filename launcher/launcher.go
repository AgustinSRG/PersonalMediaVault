// Launcher tools

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	BACKEND_BIN   = ""
	BACKUP_BIN    = ""
	FFMPEG_BIN    = ""
	FFPROBE_BIN   = ""
	FRONTEND_PATH = ""
)

type LauncherConfig struct {
	Path        string `json:"path"`
	HostName    string `json:"hostname"`
	Port        int    `json:"port"`
	Local       bool   `json:"local"`
	SSL_Cert    string `json:"ssl_cert"`
	SSL_Key     string `json:"ssl_key"`
	CacheSize   *int   `json:"cache_size"`
	LogRequests bool   `json:"log_requests"`
	Debug       bool   `json:"debug"`
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
			BACKEND_BIN = path.Join("..", "backend", getBinaryFileName("pmvd"))

			if !fileExists(BACKEND_BIN) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorBackendBin",
						Other: "Error: Could not find the backend binary (pmvd)",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "FilesMissing",
						Other: "Seems like some required files are missing.",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ReinstallFix",
						Other: "In order to fix this error, you could re-install PersonalMediaVault.",
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}
		}
	}

	// Backup tool

	BACKUP_BIN = path.Join(getDirName(), "bin", getBinaryFileName("pmv-backup"))

	if !fileExists(BACKUP_BIN) {
		BACKUP_BIN = path.Join("/usr/bin", getBinaryFileName("pmv-backup"))

		if !fileExists(BACKUP_BIN) {
			BACKUP_BIN = path.Join("..", "backup-tool", getBinaryFileName("pmv-backup"))

			if !fileExists(BACKUP_BIN) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorBackupBin",
						Other: "Error: Could not find the backup tool binary (pmv-backup)",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "FilesMissing",
						Other: "Seems like some required files are missing.",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ReinstallFix",
						Other: "In order to fix this error, you could re-install PersonalMediaVault.",
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}
		}
	}

	// FFMPEG

	FFMPEG_BIN = path.Join(getDirName(), "bin", getBinaryFileName("ffmpeg"))

	if !fileExists(FFMPEG_BIN) {

		if runtime.GOOS == "windows" {
			FFMPEG_BIN = path.Join("C:\\ffmpeg\\bin\\", getBinaryFileName("ffmpeg"))
		} else {
			FFMPEG_BIN = path.Join("/usr/bin", getBinaryFileName("ffmpeg"))
		}

		if !fileExists(FFMPEG_BIN) {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorCodecBin",
					Other: "Error: Could not find the ffmpeg binary (ffmpeg)",
				},
			})
			fmt.Println(msg)
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FilesMissing",
					Other: "Seems like some required files are missing.",
				},
			})
			fmt.Println(msg)
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ReinstallFix",
					Other: "In order to fix this error, you could re-install PersonalMediaVault.",
				},
			})
			fmt.Println(msg)
			os.Exit(1)
		}

	}

	// FFPROBE

	FFPROBE_BIN = path.Join(getDirName(), "bin", getBinaryFileName("ffprobe"))

	if !fileExists(FFPROBE_BIN) {

		if runtime.GOOS == "windows" {
			FFPROBE_BIN = path.Join("C:\\ffmpeg\\bin\\", getBinaryFileName("ffprobe"))
		} else {
			FFPROBE_BIN = path.Join("/usr/bin", getBinaryFileName("ffprobe"))
		}

		if !fileExists(FFPROBE_BIN) {
			msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ErrorProbeBin",
					Other: "Error: Could not find the ffprobe binary (ffprobe)",
				},
			})
			fmt.Println(msg)
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FilesMissing",
					Other: "Seems like some required files are missing.",
				},
			})
			fmt.Println(msg)
			msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ReinstallFix",
					Other: "In order to fix this error, you could re-install PersonalMediaVault.",
				},
			})
			fmt.Println(msg)
			os.Exit(1)
		}
	}

	// Front

	FRONTEND_PATH = path.Join(getDirName(), "www")
	if !folderExists(FRONTEND_PATH) {
		FRONTEND_PATH = "/usr/lib/pmv/www"
		if !folderExists(FRONTEND_PATH) {
			FRONTEND_PATH = "../frontend/dist"

			if !folderExists(FRONTEND_PATH) {
				msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ErrorFrontMissing",
						Other: "Error: Could not find the frontend package directory",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "FilesMissing",
						Other: "Seems like some required files are missing.",
					},
				})
				fmt.Println(msg)
				msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ReinstallFix",
						Other: "In order to fix this error, you could re-install PersonalMediaVault.",
					},
				})
				fmt.Println(msg)
				os.Exit(1)
			}
		}
	}
}

func readLauncherConfig(file string) LauncherConfig {
	b, err := os.ReadFile(file)

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
	err = json.Unmarshal(b, &config)

	if err != nil {
		return LauncherConfig{
			Port:  0,
			Local: true,
		}
	}

	return config
}

func writeLauncherConfig(file string, config LauncherConfig) error {
	// Get the json data
	jsonData, err := json.Marshal(config)

	if err != nil {
		return err
	}

	// Write file
	err = os.WriteFile(file, jsonData, FILE_PERMISSION)
	if err != nil {
		return err
	}

	return nil
}
