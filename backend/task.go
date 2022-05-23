// Media enconding tasks

package main

import (
	"io"
	"math"
	"os"
	"os/exec"
	"path"
)

func (task *ActiveTask) Run(vault *Vault) {
	if task.killed {
		return // Task killed
	}

	defer func() {
		if err := recover(); err != nil {
			switch x := err.(type) {
			case string:
				LogTaskError(task.definition.Id, "Error: "+x)
			case error:
				LogTaskError(task.definition.Id, "Error: "+x.Error())
			default:
				LogTaskError(task.definition.Id, "Task Crashed!")
			}
		}
		LogTaskDebug(task.definition.Id, "Task ended!")
	}()

	switch task.definition.Type {
	case TASK_ENCODE_ORIGINAL:
		task.RunEncodeOriginalMediaTask(vault)
	case TASK_ENCODE_RESOLUTION:
		task.RunEncodeResolutionMediaTask(vault)
	case TASK_IMAGE_PREVIEWS:
		task.RunGeneratePreviews(vault)
	}
}

// This task encodes the original media file so it's playable from the browser
// We have:
//   - Videos (mp4)
//   - Audios (mp3)
//   - Images (png)
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

	meta, err := media.ReadMeatadata(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	if !meta.OriginalReady {
		LogTaskError(task.definition.Id, "Error: Original is not ready yet.")

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

	tempFolder, err := GetTemporalFolder()

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	ext := meta.OriginalExtension

	if ext == "" {
		ext = "avi"
	}

	originalTemp := path.Join(tempFolder, "original."+ext)

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

	for finished {
		c, err := s.Read(buf)

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

		if c == 0 {
			finished = true
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

	if probe_data.Type == MediaTypeVideo {
		encoded_temp = path.Join(tempFolder, "video.mp4")
		encoded_ext = "mp4"
		cmd = MakeFFMpegEncodeToMP4OriginalCommand(originalTemp, probe_data.Format, tempFolder, userConfig)
	} else if probe_data.Type == MediaTypeAudio {
		encoded_temp = path.Join(tempFolder, "audio.mp3")
		encoded_ext = "mp3"
		cmd = MakeFFMpegEncodeToMP3Command(originalTemp, probe_data.Format, tempFolder, userConfig)
	} else if probe_data.Type == MediaTypeImage {
		encoded_temp = path.Join(tempFolder, "image.png")
		encoded_ext = "png"
		cmd = MakeFFMpegEncodeOriginalToPNGCommand(originalTemp, probe_data.Format, tempFolder, userConfig)
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

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

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

	encrypted_temp := path.Join(tempFolder, "asset.pma")

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
			return
		}

		if c == 0 {
			finished = true
			continue
		}

		err = ws.Write(buf[:c])

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(ws.file_size)))
	}

	f.Close()
	ws.Close()

	// Write changes to metadata

	task.status.SetStage("UPDATE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	metaToWrite, err := media.StartWrite(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
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
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = os.Rename(encrypted_temp, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(asset_id)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	// Save old asset to delete it
	old_asset := metaToWrite.OriginalAsset

	// Write metadata
	metaToWrite.OriginalEncoded = true
	metaToWrite.OriginalAsset = asset_id
	metaToWrite.OriginalExtension = encoded_ext
	metaToWrite.OriginalTask = 0

	err = media.EndWrite(metaToWrite, task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
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
	WipeTemporalPath(tempFolder)
}

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

	meta, err := media.ReadMeatadata(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	// check if resolution is available

	if !meta.OriginalReady {
		LogTaskError(task.definition.Id, "Error: Original is not ready yet.")

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

	tempFolder, err := GetTemporalFolder()

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.ReleaseAsset(meta.OriginalAsset)

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		return
	}

	ext := meta.OriginalExtension

	if ext == "" {
		ext = "avi"
	}

	originalTemp := path.Join(tempFolder, "original."+ext)

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

	for finished {
		c, err := s.Read(buf)

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

		if c == 0 {
			finished = true
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
		cmd = MakeFFMpegEncodeToMP4Command(originalTemp, probe_data.Format, tempFolder, &resolution, userConfig)
	} else if probe_data.Type == MediaTypeImage {
		encoded_temp = path.Join(tempFolder, "image.png")
		encoded_ext = "png"
		cmd = MakeFFMpegEncodeToPNGCommand(originalTemp, probe_data.Format, tempFolder, &resolution, userConfig)
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

	encrypted_temp := path.Join(tempFolder, "asset.pma")

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
			return
		}

		if c == 0 {
			finished = true
			continue
		}

		err = ws.Write(buf[:c])

		if err != nil {
			LogTaskError(task.definition.Id, "Error: "+err.Error())

			f.Close()
			ws.Close()

			GetVault().media.ReleaseMediaResource(task.definition.MediaId)
			WipeTemporalPath(tempFolder)
			return
		}

		bytesCopied += int64(c)
		task.status.SetProgress(float64(bytesCopied) * 100 / math.Max(1, float64(ws.file_size)))
	}

	f.Close()
	ws.Close()

	// Write changes to metadata

	task.status.SetStage("UPDATE")

	if task.killed {
		GetVault().media.ReleaseMediaResource(task.definition.MediaId)

		WipeTemporalPath(tempFolder)

		return
	}

	metaToWrite, err := media.StartWrite(task.session.key)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
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
		return
	}

	asset_lock.RequestWrite()
	asset_lock.StartWrite()

	// Move temp file
	err = os.Rename(encrypted_temp, asset_path)

	asset_lock.EndWrite()

	media.ReleaseAsset(asset_id)

	if err != nil {
		LogTaskError(task.definition.Id, "Error: "+err.Error())

		media.CancelWrite()

		GetVault().media.ReleaseMediaResource(task.definition.MediaId)
		WipeTemporalPath(tempFolder)
		return
	}

	// Write metadata
	resToWrite.Ready = true
	resToWrite.Asset = asset_id
	resToWrite.Extension = encoded_ext
	resToWrite.TaskId = 0

	err = media.EndWrite(metaToWrite, task.session.key)

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

// This task generates previews for videos
// This enables the feature that allows the user to peek images in the timeline
func (task *ActiveTask) RunGeneratePreviews(vault *Vault) {
}
