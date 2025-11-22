// Task: Encode original

package main

import (
	"io"
	"math"
	"os"
	"os/exec"
	"path"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
)

// Sets the error trace to media
// for the original encoding process
func (task *ActiveTask) SetOriginalErrorToMedia(vault *Vault, errorTrace string) {
	media := GetVault().media.AcquireMediaResource(task.definition.MediaId)

	meta, err := media.StartWrite(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if meta == nil {
		media.CancelWrite()

		LogTaskError(task.definition.Id, "Error: Media not found")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	meta.OriginalError = errorTrace

	err = media.EndWrite(meta, task.session.key, false)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	GetVault().media.ReleaseMediaResource(task.definition.MediaId)
}

// This task encodes the original media file so it's playable from the browser
// We have:
//   - Videos (mp4)
//   - Audios (mp3)
//   - Images (png)
//
// If the original file extension is already set to the expected one, no need to encode
// After encoding, the original file is replaced
func (task *ActiveTask) RunEncodeOriginalMediaTask(vault *Vault) {
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

	// Original asset must be ready for the task, not ready means broken media
	for !meta.OriginalReady {
		LogTaskError(task.definition.Id, "Error: Media not ready, but task was somehow started")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if meta.OriginalEncoded {
		// Already encoded
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

		DeleteTemporalPath(tempFolder)

		return
	}

	asset_lock.StartRead()

	s, err := encrypted_storage.CreateFileBlockEncryptReadStream(asset_path, task.session.key, FILE_PERMISSION)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		asset_lock.EndRead()

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)

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

			DeleteTemporalPath(tempFolder)

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

			DeleteTemporalPath(tempFolder)

			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(s.FileSize())))

		if task.killed {
			f.Close()
			s.Close()

			asset_lock.EndRead()

			media.ReleaseAsset(meta.OriginalAsset)

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)

			DeleteTemporalPath(tempFolder)

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

		DeleteTemporalPath(tempFolder)

		return
	}

	probe_data, err := ProbeMediaFileWithFFProbe(originalTemp)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)

		return
	}

	var encoded_temp string
	var encoded_ext string
	var cmd *exec.Cmd

	switch probe_data.Type {
	case MediaTypeVideo:
		encoded_temp = path.Join(tempFolder, "video.mp4")
		encoded_ext = "mp4"
		cmd = MakeFFMpegEncodeToMP4OriginalCommand(
			originalTemp,
			probe_data.Format,
			probe_data.Duration,
			probe_data.Width,
			probe_data.Height,
			tempFolder,
			userConfig,
			task.definition.FirstTimeEncoding && probe_data.CanCopyVideo,
			task.definition.FirstTimeEncoding && probe_data.CanCopyAudio,
		)
	case MediaTypeAudio:
		encoded_temp = path.Join(tempFolder, "audio.mp3")
		encoded_ext = "mp3"
		cmd = MakeFFMpegEncodeToMP3Command(originalTemp, probe_data.Format, tempFolder, userConfig)
	case MediaTypeImage:
		encoded_temp = path.Join(tempFolder, "image.png")
		encoded_ext = "png"
		cmd = MakeFFMpegEncodeOriginalToPNGCommand(originalTemp, probe_data.Format, tempFolder, userConfig)
	default:
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		return
	}

	// Encode media file

	task.status.SetStage("ENCODE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)

		return
	}

	err = RunFFMpegCommandAsync(cmd, probe_data.Duration, func(p float64) bool {
		task.status.SetProgress(p)

		return task.killed
	})

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)

		return
	}

	if err != nil {
		LogTaskError(task.definition.Id, err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)

		task.SetOriginalErrorToMedia(vault, err.Error())
		return
	}

	// Encrypt the file

	task.status.SetStage("ENCRYPT")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)

		return
	}

	f, err = os.OpenFile(encoded_temp, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		return
	}

	f_info, err := f.Stat()

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		return
	}

	encrypted_temp := GetTemporalFileName("pma", true)

	ws, err := encrypted_storage.CreateFileBlockEncryptWriteStream(encrypted_temp, FILE_PERMISSION)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		return
	}

	encodedSize := f_info.Size()
	err = ws.Initialize(encodedSize, ENCRYPTED_BLOCK_MAX_SIZE, task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()
		ws.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	bytesCopied = 0
	finished = false

	for !finished {
		c, err := f.Read(buf)

		if err != nil && err != io.EOF {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			DeleteTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		if err == io.EOF {
			finished = true
		}

		if c == 0 {
			continue
		}

		err = ws.Write(buf[:c])

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			DeleteTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(encodedSize)))

		if task.killed {
			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			DeleteTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}
	}

	f.Close()
	ws.Close()

	// Write changes to metadata

	task.status.SetStage("UPDATE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		DeleteTemporalPath(tempFolder)
		os.Remove(encrypted_temp)

		return
	}

	metaToWrite, err := media.StartWrite(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	asset_id := metaToWrite.NextAssetID
	metaToWrite.NextAssetID++

	found, asset_path, asset_lock = media.AcquireAsset(asset_id, ASSET_SINGLE_FILE)

	if !found {
		LogTaskError(task.definition.Id, "Error: Could not find asset to write")

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
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
		DeleteTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	// Save old asset to delete it
	old_asset := metaToWrite.OriginalAsset

	// Write metadata
	metaToWrite.OriginalEncoded = true
	metaToWrite.OriginalAsset = asset_id
	metaToWrite.OriginalExtension = encoded_ext
	metaToWrite.OriginalTask = 0
	metaToWrite.OriginalError = ""

	err = media.EndWrite(metaToWrite, task.session.key, false)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		DeleteTemporalPath(tempFolder)
		return
	}

	task.status.SetStage("FINISH")

	// Delete old asset
	found, old_asset_path, old_asset_lock := media.AcquireAsset(old_asset, ASSET_SINGLE_FILE)

	if found {
		old_asset_lock.RequestWrite()
		old_asset_lock.StartWrite()

		os.Remove(old_asset_path)

		old_asset_lock.EndWrite()

		media.ReleaseAsset(old_asset)
	}

	// Finish task
	GetVault().media.ReleaseMediaResource(task.definition.MediaId)
	DeleteTemporalPath(tempFolder)
}
