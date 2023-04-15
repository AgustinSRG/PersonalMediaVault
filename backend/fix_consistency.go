// Fix vault consistency:
// - Remove deleted media that was not removed

package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strconv"
)

// Fixes vault consistency
// vault - Reference to the vault to handle
func FixVaultConsistency(vault *Vault) {
	// Get the full list of media assets in the vault, and check if they exists in the main index

	index, err := vault.index.StartRead()

	if err != nil {
		vault.index.EndRead(index)
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
						LogInfo("Found inconsistency: Media folder prefix invalid 'media/" + dirs[i].Name() + "/" + fmt.Sprint(media_id) + "' (removing)")
						os.RemoveAll(path.Join(vault.path, "media", dirs[i].Name(), fmt.Sprint(media_id)))
						continue
					}

					exists, _, err := index.BinarySearch(media_id)

					if err != nil {
						vault.index.EndRead(index)
						LogError(err)
						os.Exit(1)
					}

					if !exists {
						// Remove directory
						LogInfo("Found inconsistency: Media folder not indexed 'media/" + dirs[i].Name() + "/" + fmt.Sprint(media_id) + "' (removing)")
						os.RemoveAll(path.Join(vault.path, "media", dirs[i].Name(), fmt.Sprint(media_id)))
					}
				}

			}
		}
	}

	vault.index.EndRead(index)
}

// Gets the media IDs from the file system
// p - Path to check
func fetchMediaIds(p string) []uint64 {
	dirs, err := os.ReadDir(p)

	if err != nil {
		LogError(err)
		return make([]uint64, 0)
	}

	result := make([]uint64, 0)

	for i := 0; i < len(dirs); i++ {
		if dirs[i].Type().IsDir() {
			mediaId, err := strconv.ParseUint(dirs[i].Name(), 10, 64)

			if err != nil {
				continue
			}

			result = append(result, mediaId)
		}
	}

	return result
}
