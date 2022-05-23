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
