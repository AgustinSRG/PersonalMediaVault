// Vault main struct

package main

import "path"

type Vault struct {
	path string

	credentials *VaultCredentialsManager // Manages vault credentials
	sessions    *SessionManager          // Manages active user sessiions

	media *MediaAssetsManager // Media assets

	tasks *TaskManager // Manages media encoding tasks

	index  *VaultMainIndex     // Main index with all media assets
	tags   *VaultTagManager    // Tags (unsorted lists)
	albums *VaultAlbumsManager // Albums (sorted lists)

	config *UserConfigManager // User config
}

func (vault *Vault) Initialize(base_path string) error {
	vault.path = base_path

	vault.credentials = &VaultCredentialsManager{}
	err := vault.credentials.Initialize(base_path)

	if err != nil {
		return err
	}

	vault.sessions = &SessionManager{}
	vault.sessions.Initialize(vault)

	vault.media = &MediaAssetsManager{}
	vault.media.Initialize(base_path)

	vault.tasks = &TaskManager{}
	err = vault.tasks.Initialize(base_path, vault)

	if err != nil {
		return err
	}

	vault.index = &VaultMainIndex{}
	err = vault.index.Initialize(path.Join(base_path, "main.index"))

	if err != nil {
		return err
	}

	vault.tags = &VaultTagManager{}
	err = vault.tags.Initialize(base_path)

	if err != nil {
		return err
	}

	vault.albums = &VaultAlbumsManager{}
	vault.albums.Initialize(base_path)

	vault.config = &UserConfigManager{}
	vault.config.Initialize(base_path)

	return nil
}