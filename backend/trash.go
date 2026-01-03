// Trash handling job

package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

const TRASH_HANDLER_LOG_PREFIX = "[TRASH HANDLER] "

// Finds files in the vault folder that should not be there
// vaultPath - The vault base path
// removeTrash - True to remove the files, false to just print them
func FindAndHandleTrashFiles(vaultPath string, removeTrash bool) {
	vaultCredentials := &VaultCredentialsManager{}
	err := vaultCredentials.Initialize(vaultPath)

	if err != nil {
		LogError(err)
		return
	}

	user := os.Getenv("VAULT_USER")
	password := os.Getenv("VAULT_PASSWORD")

	validCredentials, _ := vaultCredentials.CheckPassword(user, password)

	if !validCredentials {
		LogErrorMsg(TRASH_HANDLER_LOG_PREFIX + "Invalid vault credentials.")
		return
	}

	key, _, _, err := vaultCredentials.UnlockVault(
		os.Getenv("VAULT_USER"),
		os.Getenv("VAULT_PASSWORD"),
	)

	if err != nil {
		LogError(err)
		return
	}

	findAndHandleTrashMediaFiles(vaultPath, removeTrash, key)
	findAndHandleTrashTags(vaultPath, removeTrash, key)
	findAndHandleTrashAlbumThumbnails(vaultPath, removeTrash, key)
}

const TRASH_HANDLE_TASK_REPORT_INTERVAL = 1 * time.Second

// Finds and handles trash media files
func findAndHandleTrashMediaFiles(vaultPath string, removeTrash bool, vaultKey []byte) {
	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Checking media folders... ")

	lastTimeReportedProgress := time.Now()

	mainIndexFile := path.Join(vaultPath, "main.index")

	mainIndex, err := OpenIndexedListForReading(mainIndexFile)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogError(err)
		}

		return
	}

	defer mainIndex.Close()

	mainIndexSize, err := mainIndex.Count()

	if err != nil {
		LogError(err)
		return
	}

	dirs, err := os.ReadDir(path.Join(vaultPath, "media"))

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogError(err)
		}
		return
	}

	checkedCount := uint64(0)

	for _, dir := range dirs {
		if !dir.Type().IsDir() {
			continue
		}

		media_ids := fetchMediaIds(path.Join(vaultPath, "media", dir.Name()))

		for _, media_id := range media_ids {
			prefixByte := byte(media_id % 256)
			prefixByteHex := hex.EncodeToString([]byte{prefixByte})

			mediaPath := path.Join(vaultPath, "media", dir.Name(), fmt.Sprint(media_id))

			checkedCount += 1

			if time.Since(lastTimeReportedProgress) > TRASH_HANDLE_TASK_REPORT_INTERVAL {
				LogInfo(TRASH_HANDLER_LOG_PREFIX + "[REPORT] Folders checked: " + fmt.Sprint(media_ids) + ", vault size: " + fmt.Sprint(mainIndexSize))
				lastTimeReportedProgress = time.Now()
			}

			if dir.Name() != prefixByteHex {
				LogInfo(TRASH_HANDLER_LOG_PREFIX + "Media folder prefix invalid: " + mediaPath)

				if removeTrash {
					_ = os.RemoveAll(mediaPath)
				}

				continue
			}

			exists, _, err := mainIndex.BinarySearch(media_id)

			if err != nil {
				LogError(err)
				continue
			}

			if !exists {
				LogInfo(TRASH_HANDLER_LOG_PREFIX + "Media folder not indexed: " + mediaPath)

				if removeTrash {
					_ = os.RemoveAll(mediaPath)
				}

				continue
			}

			findAndHandleTrashMediaAssetFiles(mediaPath, removeTrash, vaultKey)
		}
	}

	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Completed media folder check. Checked folders: " + fmt.Sprint(checkedCount))
}

func findAndHandleTrashMediaAssetFiles(mediaFolder string, removeTrash bool, vaultKey []byte) {
	metadataFile := path.Join(mediaFolder, "meta.pmv")

	metadataEncrypted, err := os.ReadFile(metadataFile)

	if err != nil {
		LogErrorMsg("Error reading metadata file (" + metadataFile + "): " + err.Error())
		return
	}

	metadataDecrypted, err := encrypted_storage.DecryptFileContents(metadataEncrypted, vaultKey)

	if err != nil {
		LogErrorMsg("Error decrypting metadata file (" + metadataFile + "): " + err.Error())
		return
	}

	metadata := MediaMetadata{}

	err = json.Unmarshal(metadataDecrypted, &metadata)

	if err != nil {
		LogErrorMsg("Error reading metadata file (" + metadataFile + "): " + err.Error())
		return
	}

	assetIdsInUse := metadata.GetUsedAssetIdSet()

	files, err := os.ReadDir(mediaFolder)

	if err != nil {
		LogErrorMsg("Error reading directory (" + mediaFolder + "): " + err.Error())
		return
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		fileName := file.Name()

		if !strings.HasSuffix(fileName, ".pma") || (!strings.HasPrefix(fileName, "s_") && !strings.HasPrefix(fileName, "m_")) {
			continue
		}

		fileNameParts := strings.Split(fileName, ".")

		if len(fileNameParts) != 2 {
			continue
		}

		fileNameParts = strings.Split(fileNameParts[0], "_")

		if len(fileNameParts) != 2 {
			continue
		}

		assetIdNum, err := strconv.ParseUint(fileNameParts[1], 10, 64)

		if err != nil {
			continue
		}

		if assetIdsInUse[assetIdNum] {
			continue
		}

		filePath := path.Join(mediaFolder, fileName)

		LogInfo(TRASH_HANDLER_LOG_PREFIX + "Found orphan media asset file: " + filePath)

		if removeTrash {
			_ = os.Remove(filePath)
		}
	}
}

func findAndHandleTrashAlbumThumbnails(vaultPath string, removeTrash bool, vaultKey []byte) {
	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Checking album thumbnails...")

	albumsFile := path.Join(vaultPath, "albums.pmv")

	albumDataEncrypted, err := os.ReadFile(albumsFile)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogErrorMsg("Error reading file (" + albumsFile + "): " + err.Error())
		}
		return
	}

	albumDataDecrypted, err := encrypted_storage.DecryptFileContents(albumDataEncrypted, vaultKey)

	if err != nil {
		LogErrorMsg("Error decrypting file (" + albumsFile + "): " + err.Error())
		return
	}

	albumData := VaultAlbumsData{}

	err = json.Unmarshal(albumDataDecrypted, &albumData)

	if err != nil {
		LogErrorMsg("Error reading file (" + albumsFile + "): " + err.Error())
		return
	}

	thumbnailIdMap := albumData.GetThumbnailsIdMap()

	albumThumbnailFolder := path.Join(vaultPath, "thumb_album")

	files, err := os.ReadDir(albumThumbnailFolder)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogError(err)
		}
		return
	}

	checkedCount := uint64(0)

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		fileName := file.Name()

		if !strings.HasPrefix(fileName, "s_") || !strings.HasSuffix(fileName, ".pma") {
			continue
		}

		fileNameParts := strings.Split(fileName, ".")

		if len(fileNameParts) != 2 {
			continue
		}

		fileNameParts = strings.Split(fileNameParts[0], "_")

		if len(fileNameParts) != 2 {
			continue
		}

		assetIdNum, err := strconv.ParseUint(fileNameParts[1], 10, 64)

		if err != nil {
			continue
		}

		checkedCount++

		if thumbnailIdMap[assetIdNum] {
			continue
		}

		filePath := path.Join(albumThumbnailFolder, fileName)

		LogInfo(TRASH_HANDLER_LOG_PREFIX + "Found orphan album thumbnail file: " + filePath)

		if removeTrash {
			_ = os.Remove(filePath)
		}
	}

	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Completed album thumbnails check. Checked files: " + fmt.Sprint(checkedCount))
}

func findAndHandleTrashTags(vaultPath string, removeTrash bool, vaultKey []byte) {
	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Checking tags...")

	tagsFile := path.Join(vaultPath, "tag_list.pmv")

	tagDataEncrypted, err := os.ReadFile(tagsFile)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogErrorMsg("Error reading file (" + tagsFile + "): " + err.Error())
		}
		return
	}

	tagDataDecrypted, err := encrypted_storage.DecryptFileContents(tagDataEncrypted, vaultKey)

	if err != nil {
		LogErrorMsg("Error decrypting file (" + tagsFile + "): " + err.Error())
		return
	}

	tagData := VaultTagListData{}

	err = json.Unmarshal(tagDataDecrypted, &tagData)

	if err != nil {
		LogErrorMsg("Error reading file (" + tagsFile + "): " + err.Error())
		return
	}

	tagIdMap := tagData.GetExistingTagsMap()

	tagsFolder := path.Join(vaultPath, "tags")

	files, err := os.ReadDir(tagsFolder)

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			LogError(err)
		}
		return
	}

	checkedCount := uint64(0)

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		fileName := file.Name()

		if !strings.HasPrefix(fileName, "tag_") || !strings.HasSuffix(fileName, ".index") {
			continue
		}

		fileNameParts := strings.Split(fileName, ".")

		if len(fileNameParts) != 2 {
			continue
		}

		fileNameParts = strings.Split(fileNameParts[0], "_")

		if len(fileNameParts) != 2 {
			continue
		}

		tagId, err := strconv.ParseUint(fileNameParts[1], 10, 64)

		if err != nil {
			continue
		}

		checkedCount++

		if tagIdMap[tagId] {
			continue
		}

		filePath := path.Join(tagsFolder, fileName)

		LogInfo(TRASH_HANDLER_LOG_PREFIX + "Found orphan tag file: " + filePath)

		if removeTrash {
			_ = os.Remove(filePath)
		}
	}

	LogInfo(TRASH_HANDLER_LOG_PREFIX + "Completed tags check. Checked files: " + fmt.Sprint(checkedCount))
}

// Gets the media IDs from the file system
// p - Path to check
func fetchMediaIds(p string) []uint64 {
	dirs, err := os.ReadDir(p)

	if err != nil {
		LogError(err)
		return make([]uint64, 0)
	}

	result := make([]uint64, 0)

	for _, dir := range dirs {
		if dir.Type().IsDir() {
			mediaId, err := strconv.ParseUint(dir.Name(), 10, 64)

			if err != nil {
				continue
			}

			result = append(result, mediaId)
		}
	}

	return result
}
