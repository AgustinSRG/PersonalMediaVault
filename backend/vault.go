// Vault main struct

package main

import "path"

// Vault main data struct
type Vault struct {
	path string // Vault path

	credentials *VaultCredentialsManager // Manages vault credentials
	sessions    *SessionManager          // Manages active user sessions
	invites     *InvitationManager       // Manages invite codes

	media *MediaAssetsManager // Media assets

	tasks *TaskManager // Manages media encoding tasks

	index  *VaultMainIndex     // Main index with all media assets
	tags   *VaultTagManager    // Tags (unsorted lists)
	albums *VaultAlbumsManager // Albums (sorted lists)

	homePage *HomePageConfigManager // Home page config

	config *UserConfigManager // User config
}

// Vault initialization process
// base_path - Vault path
// preview_cache_size - Max size of the preview cache
func (vault *Vault) Initialize(base_path string, preview_cache_size int) error {
	vault.path = base_path

	vault.credentials = &VaultCredentialsManager{}
	err := vault.credentials.Initialize(base_path)

	if err != nil {
		return err
	}

	vault.sessions = &SessionManager{}
	vault.sessions.Initialize(vault)

	vault.invites = &InvitationManager{}
	vault.invites.Initialize(vault)

	vault.media = &MediaAssetsManager{}
	vault.media.Initialize(base_path, preview_cache_size)

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

	vault.homePage = &HomePageConfigManager{}
	vault.homePage.Initialize(base_path)

	return nil
}
