// Asset deletion utils

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// Asset deletion manager
type AssetDeletionManager struct {
	path string // Vault base path
}

// Name of the folder where the pending deletion marks are stored
const PENDING_DELETION_MARKS_FOLDER_NAME = "pending_delete"

// Suffix to append to deletion markers
const PENDING_DELETION_MARK_SUFFIX = ".pending.delete.tmp"

// Prefix for marks on media assets (single files)
const PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_SINGLE = "ms"

// Prefix for marks on media assets (multi files)
const PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_MULTI = "mm"

// Prefix for marks on album thumbnails
const PENDING_DELETION_MARK_TYPE_ALBUM_THUMBNAIL = "at"

// Initializes the deletion manager
// Deletes any pending assets that may need deletion
// base_path - vault base path
func (m *AssetDeletionManager) Initialize(base_path string) {
	m.path = base_path

	m.deletePendingAssets()
}

// Ensures directory to store the deletion marks exists
func (m *AssetDeletionManager) ensureMarksFolderExists() string {
	pendingDeletionMarksPath := path.Join(m.path, PENDING_DELETION_MARKS_FOLDER_NAME)

	info, err := os.Stat(pendingDeletionMarksPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Folder not found, create it
			err = os.MkdirAll(pendingDeletionMarksPath, FOLDER_PERMISSION)

			if err != nil {
				LogErrorMsg(fmt.Sprintf("Failed to create '%s': %s", pendingDeletionMarksPath, err.Error()))
				LogErrorMsg("The directory for pending deletion marks is inaccessible. This may case future errors when deleting assets.")
			}
		} else {
			LogErrorMsg(fmt.Sprintf("Failed to check directory '%s': %s", pendingDeletionMarksPath, err.Error()))
		}

		return pendingDeletionMarksPath
	}

	if !info.IsDir() {
		LogWarning(fmt.Sprintf("The directory to store pending deletion marks ('%s') exists as a non-directory. Attempting deletion...", pendingDeletionMarksPath))

		err = os.Remove(pendingDeletionMarksPath)

		if err != nil {
			LogErrorMsg(fmt.Sprintf("Failed to delete '%s': %s", pendingDeletionMarksPath, err.Error()))
			LogErrorMsg("The directory for pending deletion marks is inaccessible. This may case future errors when deleting assets.")

			return pendingDeletionMarksPath
		}

		err = os.MkdirAll(pendingDeletionMarksPath, FOLDER_PERMISSION)

		if err != nil {
			LogErrorMsg(fmt.Sprintf("Failed to create '%s': %s", pendingDeletionMarksPath, err.Error()))
			LogErrorMsg("The directory for pending deletion marks is inaccessible. This may case future errors when deleting assets.")
		}
	}

	return pendingDeletionMarksPath
}

// Deletes any pending assets
// Call only once on initialization
func (m *AssetDeletionManager) deletePendingAssets() {
	pendingDeletionMarksPath := m.ensureMarksFolderExists()

	entries, err := os.ReadDir(pendingDeletionMarksPath)

	if err != nil {
		LogErrorMsg(fmt.Sprintf("Failed to read directory '%s': %s", pendingDeletionMarksPath, err.Error()))

		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		entryName := entry.Name()

		if !strings.HasSuffix(entryName, PENDING_DELETION_MARK_SUFFIX) {
			continue
		}

		m.handlePendingDeletionMark(pendingDeletionMarksPath, entryName)
	}
}

// Handles a deletion mark (called during initialization)
func (m *AssetDeletionManager) handlePendingDeletionMark(pendingDeletionMarksPath string, markName string) {
	parts := strings.Split(markName, ".")

	if len(parts) < 1 {
		return
	}

	subParts := strings.Split(parts[0], "-")

	if len(subParts) < 1 {
		return
	}

	assetType := subParts[0]

	switch assetType {
	case PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_SINGLE:
		{
			if len(subParts) != 3 {
				return
			}

			media_id, err := strconv.ParseUint(subParts[1], 10, 64)

			if err != nil {
				return
			}

			asset_id, err := strconv.ParseUint(subParts[2], 10, 64)

			if err != nil {
				return
			}

			assetPath := GetMediaAssetPath(ResolveMediaPath(m.path, media_id), asset_id, ASSET_SINGLE_FILE)

			err = RemoveFileIfExists(assetPath)

			if err != nil {
				LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", assetPath, err.Error()))
				return
			} else {
				LogInfo(fmt.Sprintf("Deleted file pending for deletion in the previous run: '%s'", assetPath))
			}
		}
	case PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_MULTI:
		{
			if len(subParts) != 3 {
				return
			}

			media_id, err := strconv.ParseUint(subParts[1], 10, 64)

			if err != nil {
				return
			}

			asset_id, err := strconv.ParseUint(subParts[2], 10, 64)

			if err != nil {
				return
			}

			assetPath := GetMediaAssetPath(ResolveMediaPath(m.path, media_id), asset_id, ASSET_MULTI_FILE)

			err = RemoveFileIfExists(assetPath)

			if err != nil {
				LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", assetPath, err.Error()))
				return
			} else {
				LogInfo(fmt.Sprintf("Deleted file pending for deletion in the previous run: '%s'", assetPath))
			}
		}
	case PENDING_DELETION_MARK_TYPE_ALBUM_THUMBNAIL:
		{
			if len(subParts) != 2 {
				return
			}

			asset_id, err := strconv.ParseUint(subParts[1], 10, 64)

			if err != nil {
				return
			}

			assetPath := GetThumbnailAssetPath(m.path, asset_id)

			err = RemoveFileIfExists(assetPath)

			if err != nil {
				LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", assetPath, err.Error()))
				return
			} else {
				LogInfo(fmt.Sprintf("Deleted file pending for deletion in the previous run: '%s'", assetPath))
			}
		}
	default:
		return
	}

	markPath := path.Join(pendingDeletionMarksPath, markName)

	err := os.Remove(markPath)

	if err != nil {
		LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", markPath, err.Error()))
	}
}

// Creates an empty pending deletion marker file
// Best effort: On error, just log it, do not block the process
func createPendingDeletionMarkFile(markPath string) {
	file, err := os.Create(markPath)

	if err != nil {
		LogErrorMsg(fmt.Sprintf("Failed to create mark file '%s': %s", markPath, err.Error()))
	} else {
		file.Close()
	}
}

// Creates the marker file for a pending deletion
// Best effort: On error, just log it, do not block the process
func removePendingDeletionMarkFile(markPath string) {
	err := RemoveFileIfExists(markPath)

	if err != nil {
		LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", markPath, err.Error()))
	}
}

// Deletes album thumbnail asset (background process)
// asset_id - Asset ID
// markPath - Path to the mark file, in order to to delete it if deletion is successful
func (m *AssetDeletionManager) deleteAlbumThumbnailAssetBackground(asset_id uint64, markPath string) {
	success, asset_path, asset_lock := GetVault().albums.AcquireThumbnailAsset(asset_id)

	if success {
		asset_lock.RequestWrite()
		asset_lock.StartWrite()

		err := RemoveFileIfExists(asset_path)

		if err != nil {
			LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", asset_path, err.Error()))
		} else {
			// Success, remove mark
			removePendingDeletionMarkFile(markPath)
		}

		asset_lock.EndWrite()

		GetVault().albums.ReleaseThumbnailAsset(asset_id)
	}
}

// Deletes album thumbnail asset
// asset_id - Asset ID
func (m *AssetDeletionManager) DeleteAlbumThumbnailAsset(asset_id uint64) {
	pendingDeletionMarksPath := m.ensureMarksFolderExists()

	markPath := path.Join(pendingDeletionMarksPath, PENDING_DELETION_MARK_TYPE_ALBUM_THUMBNAIL+"-"+fmt.Sprint(asset_id)+PENDING_DELETION_MARK_SUFFIX)

	createPendingDeletionMarkFile(markPath)

	go m.deleteAlbumThumbnailAssetBackground(asset_id, markPath)
}

// Deletes media asset
// This function runs as a background process
// media_id - Media ID
// asset_id - Asset ID
// asset_type - Asset type
// markPath - Path to the mark file, in order to to delete it if deletion is successful
func (m *AssetDeletionManager) deleteMediaAssetFileBackground(media_id uint64, asset_id uint64, asset_type string, markPath string) {
	media := GetVault().media.AcquireMediaResource(media_id)

	if media == nil {
		// Media not found

		removePendingDeletionMarkFile(markPath)
		return
	}

	success, asset_path, asset_lock := media.AcquireAsset(asset_id, asset_type)

	if !success {
		// Media was deleted

		GetVault().media.ReleaseMediaResource(media_id)

		removePendingDeletionMarkFile(markPath)

		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	err := RemoveFileIfExists(asset_path)

	if err != nil {
		LogErrorMsg(fmt.Sprintf("Failed to delete file '%s': %s", asset_path, err.Error()))
	} else {
		// Success, remove mark
		removePendingDeletionMarkFile(markPath)
	}

	asset_lock.EndWrite()

	media.ReleaseAsset(asset_id)

	GetVault().media.ReleaseMediaResource(media_id)
}

func getMarkPrefixByMediaAssetType(asset_type string) string {
	switch asset_type {
	case ASSET_MULTI_FILE:
		return PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_MULTI
	default:
		return PENDING_DELETION_MARK_TYPE_MEDIA_ASSET_SINGLE
	}
}

// Deletes media asset (single file)
// media_id - Media ID
// asset_id - Asset ID
// asset_type - Asset type
func (m *AssetDeletionManager) DeleteMediaAsset(media_id uint64, asset_id uint64, asset_type string) {
	pendingDeletionMarksPath := m.ensureMarksFolderExists()

	markPath := path.Join(pendingDeletionMarksPath, getMarkPrefixByMediaAssetType(asset_type)+"-"+fmt.Sprint(media_id)+"-"+fmt.Sprint(asset_id)+PENDING_DELETION_MARK_SUFFIX)

	createPendingDeletionMarkFile(markPath)

	go m.deleteMediaAssetFileBackground(media_id, asset_id, asset_type, markPath)
}
