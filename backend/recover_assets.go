// Recover assets:
// - Recovers non-indexed media assets

package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
)

// Recovers non-indexed media assets
// vault - Reference to the vault to handle
func RecoverVaultAssets(vault *Vault) {
	// Get the full list of media assets in the vault, and check if they exists in the main index

	index, err := vault.index.StartWrite()

	if err != nil {
		vault.index.CancelWrite(index)
		LogError(err)
		os.Exit(1)
	}

	dirs, err := os.ReadDir(path.Join(vault.path, "media"))

	if err == nil {
		for i := 0; i < len(dirs); i++ {
			if dirs[i].Type().IsDir() {
				media_ids := fetchMediaIds(path.Join(vault.path, "media", dirs[i].Name()))

				for j := 0; j < len(media_ids); j++ {
					media_id := media_ids[j]

					prefixByte := byte(media_id % 256)
					prefixByteHex := hex.EncodeToString([]byte{prefixByte})

					if dirs[i].Name() != prefixByteHex {
						continue
					}

					exists, _, err := index.file.BinarySearch(media_id)

					if err != nil {
						vault.index.CancelWrite(index)
						LogError(err)
						os.Exit(1)
					}

					if !exists {
						// Remove directory
						LogInfo("Found missing asset: Media folder not indexed 'media/" + dirs[i].Name() + "/" + fmt.Sprint(media_id) + "' (adding)")
						_, _, err = index.file.AddValue(media_id)

						if err != nil {
							vault.index.CancelWrite(index)
							LogError(err)
							os.Exit(1)
						}
					}
				}

			}
		}
	}

	err = vault.index.EndWrite(index)

	if err != nil {
		LogError(err)
		os.Exit(1)
	}
}
