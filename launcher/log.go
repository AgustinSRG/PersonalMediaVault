// Log files

package main

import (
	"fmt"
	"os"
	"path"
	"time"
)

const LIMIT_LOG_FILES = 100

func getLogsPath() (string, error) {
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	logsFolder := path.Join(userCacheDir, "PersonalMediaVault", "logs")

	return logsFolder, nil
}

func getLogFileName() (string, error) {
	userCacheDir, err := os.UserCacheDir()

	if err != nil {
		return "", err
	}

	logsFolder := path.Join(userCacheDir, "PersonalMediaVault", "logs")

	if !folderExists(logsFolder) {
		err = os.MkdirAll(logsFolder, FOLDER_PERMISSION)

		if err != nil {
			return "", err
		}
	}

	// Remove files if there are too many

	files, err := os.ReadDir(logsFolder)

	if err != nil {
		return "", err
	}

	if len(files) > LIMIT_LOG_FILES {
		for i := 0; i < (len(files) - LIMIT_LOG_FILES); i++ {
			os.Remove(path.Join(logsFolder, files[i].Name()))
		}
	}

	// PID

	pid := fmt.Sprint(os.Getpid())

	// Date

	year := fmt.Sprint(time.Now().UTC().Year())
	month := fmt.Sprint(int(time.Now().UTC().Month()))
	if len(month) < 2 {
		month = "0" + month
	}
	day := fmt.Sprint(time.Now().UTC().Day())
	if len(day) < 2 {
		day = "0" + day
	}

	ts := fmt.Sprint(time.Now().UnixMilli())

	// Make filename

	fileName := year + "-" + month + "-" + day + "-" + ts + "-" + pid + ".log"

	return path.Join(logsFolder, fileName), nil
}
