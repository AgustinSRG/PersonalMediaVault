// Backup tool for PMV

package main

import (
	"fmt"
	"os"
	"path"
)

func main() {
	// Parse arguments
	args := os.Args

	if len(args) != 3 {
		fmt.Println("Backup tool for Personal Media Vault.")
		fmt.Println("Usage: pmv-backup </path/to/vault> </path/to/backup>")
		return
	}

	vaultPath := args[1]
	backupPath := args[2]

	if !CheckFileExists(vaultPath) {
		fmt.Println("Path does not exists: " + vaultPath)
		os.Exit(1)
	}

	if !CheckFileExists(backupPath) {
		fmt.Println("Path does not exists: " + backupPath)
		os.Exit(1)
	}

	fmt.Println("Backup tool for Personal Media Vault.")

	fmt.Println("Fetching metadata from: " + vaultPath)

	if !CheckFileExists(path.Join(vaultPath, "credentials.json")) {
		fmt.Println("Could not find a vault in the specified path")
		os.Exit(1)
	}

	totalEntries := make([]BackupEntry, 0)

	mediaFiles := findBackupEntries(vaultPath, backupPath, "./media")
	totalEntries = append(totalEntries, mediaFiles...)

	tagFiles := findBackupEntries(vaultPath, backupPath, "./tags")
	totalEntries = append(totalEntries, tagFiles...)

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "main.index"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "credentials.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "media_ids.json"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tasks.json"))

	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "albums.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "tag_list.pmv"))
	totalEntries = append(totalEntries, makeBackupEntry(vaultPath, backupPath, "./", "user_config.pmv"))

	fmt.Println("Initializing backup...")

	progressInt := int64(0)
	prevProgress := int64(0)

	statFilesCopied := 0

	fmt.Print("Making backup...")

	for i := 0; i < len(totalEntries); i++ {

		c := backupFile(totalEntries[i])

		if c {
			statFilesCopied++
		}

		progressInt = int64(i+1) * 100 / int64(len(totalEntries))
		if prevProgress != progressInt {
			prevProgress = progressInt
			fmt.Print("\rMaking backup... (" + fmt.Sprint(prevProgress) + "%)")
		}
	}

	fmt.Print("\n")
	fmt.Println("Backup done. Total files copied: " + fmt.Sprint(statFilesCopied))
}
