// Configuration file finder

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path"
)

func vaultPathHashTag(vaultPath string) string {
	hasher := sha256.New()
	hasher.Write([]byte(vaultPath))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)[31:]
}

func vaultPathConfigFile(vaultPath string) (string, error) {
	userConfigDir, err := os.UserConfigDir()

	if err != nil {
		return "", err
	}

	vaultConfigDir := path.Join(userConfigDir, "PersonalMediaVault", "launcher_config")

	err = os.MkdirAll(vaultConfigDir, FOLDER_PERMISSION)

	if err != nil {
		return "", err
	}

	pathHashTag := vaultPathHashTag(vaultPath)

	return path.Join(vaultConfigDir, pathHashTag+".json"), nil
}

func getLauncherConfigFile(vaultPath string) string {
	generalLauncherConfigFile := path.Join(vaultPath, "launcher.config.json")
	specificLauncherConfigFile, err := vaultPathConfigFile(vaultPath)

	if err != nil {
		return generalLauncherConfigFile
	}

	if !fileExists(specificLauncherConfigFile) && fileExists(generalLauncherConfigFile) {
		_, err = CopyFile(generalLauncherConfigFile, specificLauncherConfigFile)

		if err != nil {
			return specificLauncherConfigFile
		}
	}

	return specificLauncherConfigFile
}
