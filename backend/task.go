// Media enconding tasks

package main

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
}

// This task encodes the media file to a specific resolution
// Only available for images and for videos
func (task *ActiveTask) RunEncodeResolutionMediaTask(vault *Vault) {
}

// This task generates previews for videos
// This enables the feature that allows the user to peek images in the timeline
func (task *ActiveTask) RunGeneratePreviews(vault *Vault) {
}
