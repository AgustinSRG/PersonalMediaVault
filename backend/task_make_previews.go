// Task: Make previews

package main

import (
	"io"
	"math"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

type VideoPreviewTempFile struct {
	file  string
	index int64
}

func getPreviewsTempFileIndex(name string) (bool, int64) {
	parts := strings.Split(name, ".")

	if len(parts) != 2 {
		return false, 0
	}

	parts = strings.Split(parts[0], "_")

	if len(parts) != 2 {
		return false, 0
	}

	i, err := strconv.ParseInt(parts[1], 10, 64)

	if err != nil {
		return false, 0
	}

	return true, i
}

// This task generates previews for videos
// This enables the feature that allows the user to peek images in the timeline
func (task *ActiveTask) RunGeneratePreviews(vault *Vault) {
	// Read media metadata

	task.status.SetStage("PREPARE")

	userConfig, err := GetVault().config.Read(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())
		return
	}

	media := GetVault().media.AcquireMediaResource(task.definition.MediaId)

	meta, err := media.ReadMetadata(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if meta == nil {
		LogTaskError(task.definition.Id, "Error: Media not found")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	// Check metadata

	// Original asset must be ready for the task, not ready means broken media
	for !meta.OriginalReady {
		LogTaskError(task.definition.Id, "Error: Media not ready, but task was somehow started")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if meta.PreviewsReady {
		// Previews are already encoded
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	found, asset_path, asset_lock := media.AcquireAsset(meta.OriginalAsset, ASSET_SINGLE_FILE)

	if !found {
		LogTaskError(task.definition.Id, "Error: Original not found.")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	// Copy the original to a temp file and decrypt it, so ffmpeg can actually read it

	task.status.SetStage("COPY")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	tempFolder, err := GetTemporalFolder(false)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	originalTemp := path.Join(tempFolder, "original")

	f, err := os.OpenFile(originalTemp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FILE_PERMISSION)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	asset_lock.StartRead()

	s, err := CreateFileBlockEncryptReadStream(asset_path, task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		asset_lock.EndRead()

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	buf := make([]byte, 1024*1024)

	var finished bool = false
	var bytesCopied int64 = 0

	for !finished {
		c, err := s.Read(buf)

		if err != nil && err != io.EOF {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			s.Close()

			asset_lock.EndRead()

			media.ReleaseAsset(meta.OriginalAsset)

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)

			WipeTemporalPath(tempFolder)

			return
		}

		if err == io.EOF {
			finished = true
		}

		if c == 0 {
			continue
		}

		_, err = f.Write(buf[:c])

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			s.Close()

			asset_lock.EndRead()

			media.ReleaseAsset(meta.OriginalAsset)

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)

			WipeTemporalPath(tempFolder)

			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(s.file_size)))

		if task.killed {
			f.Close()
			s.Close()

			asset_lock.EndRead()

			media.ReleaseAsset(meta.OriginalAsset)

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)

			WipeTemporalPath(tempFolder)

			return
		}
	}

	f.Close()
	s.Close()

	asset_lock.EndRead()

	media.ReleaseAsset(meta.OriginalAsset)

	// Original is now decrypted
	// Get probe data

	task.status.SetStage("PROBE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	probe_data, err := ProbeMediaFileWithFFProbe(originalTemp)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	var cmd *exec.Cmd
	var intervalSeconds int32

	if probe_data.Type == MediaTypeVideo {
		cmd, intervalSeconds = MakeFFMpegEncodeToPreviewsCommand(originalTemp, probe_data.Format, probe_data.Duration, tempFolder, userConfig)
	} else {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	// Encode previews

	task.status.SetStage("ENCODE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	err = RunFFMpegCommandAsync(cmd, probe_data.Duration, func(p float64) bool {
		task.status.SetProgress(p)

		return task.killed
	})

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	// Encrypt the previews

	task.status.SetStage("ENCRYPT")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	files, err := os.ReadDir(tempFolder)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	imagesSorted := make([]VideoPreviewTempFile, 0)

	for i := 0; i < len(files); i++ {
		if !files[i].Type().IsRegular() {
			continue
		}

		if !strings.HasSuffix(files[i].Name(), ".jpg") {
			continue
		}

		file_path := path.Join(tempFolder, files[i].Name())
		file_valid, file_index := getPreviewsTempFileIndex(files[i].Name())

		if !file_valid {
			continue
		}

		imagesSorted = append(imagesSorted, VideoPreviewTempFile{
			file:  file_path,
			index: file_index,
		})
	}

	sort.SliceStable(imagesSorted, func(i, j int) bool {
		return imagesSorted[i].index < imagesSorted[j].index
	})

	if len(imagesSorted) == 0 {
		LogTaskError(task.definition.Id, "Error: No preview images found")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	encrypted_temp := GetTemporalFileName("pma", true)

	ws, err := CreateMultiFilePackWriteStream(encrypted_temp)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	err = ws.Initialize(int64(len(imagesSorted)))

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		ws.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	for i := 0; i < len(imagesSorted); i++ {
		// Read image file
		bytesImage, err := os.ReadFile(imagesSorted[i].file)

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		// Encrypt data

		encData, err := encryptFileContents(bytesImage, AES256_ZIP, task.session.key)

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		// Write data (pack)

		err = ws.PutFile(encData)

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		task.status.SetProgress(float64(i) * 100 / float64(len(imagesSorted)))

		if task.killed {
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}
	}

	ws.Close()

	// Write changes to metadata

	task.status.SetStage("UPDATE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)

		return
	}

	metaToWrite, err := media.StartWrite(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	asset_id := metaToWrite.NextAssetID
	metaToWrite.NextAssetID++

	found, asset_path, asset_lock = media.AcquireAsset(asset_id, ASSET_MULTI_FILE)

	if !found {
		LogTaskError(task.definition.Id, "Error: Could not find asset to write")

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = RenameAndReplace(encrypted_temp, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(asset_id)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	// Write metadata
	metaToWrite.PreviewsReady = true
	metaToWrite.PreviewsAsset = asset_id
	metaToWrite.PreviewsInterval = float64(intervalSeconds)
	metaToWrite.PreviewsTask = 0

	err = media.EndWrite(metaToWrite, task.session.key, false)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	task.status.SetStage("FINISH")

	// Finish task
	GetVault().media.ReleaseMediaResource(task.definition.MediaId)
	WipeTemporalPath(tempFolder)
}
