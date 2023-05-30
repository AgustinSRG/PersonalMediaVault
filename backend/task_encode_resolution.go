// Task: Encode resolution

package main

import (
	"io"
	"math"
	"os"
	"os/exec"
	"path"
)

// This task encodes the media file to a specific resolution
// Only available for images and for videos
func (task *ActiveTask) RunEncodeResolutionMediaTask(vault *Vault) {
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

	// check if resolution is available

	// Original asset must be ready for the task, not ready means broken media
	for !meta.OriginalReady {
		LogTaskError(task.definition.Id, "Error: Media not ready, but task was somehow started")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if meta.Resolutions == nil {
		LogTaskError(task.definition.Id, "Error: No resolutions available.")

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	var foundResolution = false

	for i := 0; i < len(meta.Resolutions); i++ {
		if meta.Resolutions[i].Width == task.definition.Resolution.Width && meta.Resolutions[i].Height == task.definition.Resolution.Height && meta.Resolutions[i].Fps == task.definition.Resolution.Fps {

			foundResolution = true

			if meta.Resolutions[i].Ready {
				// Resolution is already done, no need to continue
				GetVault().media.ReleaseMediaResource(task.definition.MediaId)

				return
			}

			break
		}
	}

	if !foundResolution {
		LogTaskError(task.definition.Id, "Error: No resolution available.")

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

	var encoded_temp string
	var encoded_ext string
	var cmd *exec.Cmd

	resolution := UserConfigResolution{
		Width:  task.definition.Resolution.Width,
		Height: task.definition.Resolution.Height,
		Fps:    task.definition.Resolution.Fps,
	}

	if probe_data.Type == MediaTypeVideo {
		encoded_temp = path.Join(tempFolder, "video.mp4")
		encoded_ext = "mp4"
		cmd = MakeFFMpegEncodeToMP4Command(originalTemp, probe_data.Format, probe_data.Duration, tempFolder, &resolution, probe_data.Width, probe_data.Height, userConfig)
	} else if probe_data.Type == MediaTypeImage {
		encoded_temp = path.Join(tempFolder, "image.png")
		encoded_ext = "png"
		cmd = MakeFFMpegEncodeToPNGCommand(originalTemp, probe_data.Format, tempFolder, &resolution, probe_data.Width, probe_data.Height, userConfig)
	} else {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	// Encode media file

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

	// Encrypt the file

	task.status.SetStage("ENCRYPT")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	f, err = os.OpenFile(encoded_temp, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	f_info, err := f.Stat()

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	encrypted_temp := GetTemporalFileName("pma", true)

	ws, err := CreateFileBlockEncryptWriteStream(encrypted_temp)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	err = ws.Initialize(f_info.Size(), task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		f.Close()
		ws.Close()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
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
			WipeTemporalPath(tempFolder)
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
			WipeTemporalPath(tempFolder)
			os.Remove(encrypted_temp)
			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(ws.file_size)))

		if task.killed {
			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
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

	var resToWrite *MediaResolution
	foundResolution = false

	for i := 0; i < len(metaToWrite.Resolutions); i++ {
		if metaToWrite.Resolutions[i].Width == task.definition.Resolution.Width && meta.Resolutions[i].Height == task.definition.Resolution.Height && meta.Resolutions[i].Fps == task.definition.Resolution.Fps {

			foundResolution = true
			resToWrite = &metaToWrite.Resolutions[i]

			break
		}
	}

	if !foundResolution {
		LogTaskError(task.definition.Id, "Error: Could not find resolution in metadata")

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		os.Remove(encrypted_temp)
		return
	}

	asset_id := metaToWrite.NextAssetID
	metaToWrite.NextAssetID++

	found, asset_path, asset_lock = media.AcquireAsset(asset_id, ASSET_SINGLE_FILE)

	if !found {
		LogTaskError(task.definition.Id, "Error: Cound not find asset to write")

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
	resToWrite.Ready = true
	resToWrite.Asset = asset_id
	resToWrite.Extension = encoded_ext
	resToWrite.TaskId = 0

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
