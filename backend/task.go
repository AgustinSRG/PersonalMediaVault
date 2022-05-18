// Media enconding tasks

package main

func (task *ActiveTask) Run(vault *Vault) {
	if task.killed {
		return // Task killed
	}
}
