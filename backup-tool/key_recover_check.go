// Checks for key recovery

package main

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"

	encrypted_storage "github.com/AgustinSRG/encrypted-storage"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Minified version of media metadata
type MediaMetadataMin struct {
	Id    uint64 `json:"id"`    // Media ID
	Title string `json:"title"` // Title
}

// Tags data
type VaultTagListData struct {
	NextId uint64            `json:"next_id"` // ID for the next tag to add
	Tags   map[uint64]string `json:"tags"`    // Tags Map (id -> name)
}

// Album data
type VaultAlbumData struct {
	Name         string   `json:"name"`  // Name of the album
	List         []uint64 `json:"list"`  // Ordered list of media to play
	LastModified int64    `json:"lm"`    // Last modified timestamp
	Thumbnail    *uint64  `json:"thumb"` // Thumbnail asset
}

// Album list data
type VaultAlbumsData struct {
	NextId               uint64                     `json:"next_id"`                 // Id for the next album to create
	NextThumbnailAssetId uint64                     `json:"next_thumb_id,omitempty"` // Id for the next album thumbnail asset
	Albums               map[uint64]*VaultAlbumData `json:"albums"`                  // Albums (Id -> Data)
}

func checkKeyForRecoveryCheckTags(vaultPath string, key []byte) error {
	filePath := path.Join(vaultPath, "tag_list.pmv")

	if !CheckFileExists(filePath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoTags",
				Other: "This vault does not have any tags.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		return nil
	}

	// Load file
	b, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	bJSON, err := encrypted_storage.DecryptFileContents(b, key)

	if err != nil {
		return err
	}

	var mp VaultTagListData

	err = json.Unmarshal(bJSON, &mp)

	if err != nil {
		return err
	}

	if len(mp.Tags) == 0 {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoTags",
				Other: "This vault does not have any tags.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultPrintTags",
				Other: "Vault tags (Total: {{.TagCount}}):",
			},
			TemplateData: map[string]interface{}{
				"TagCount": len(mp.Tags),
			},
		})
		fmt.Fprintln(os.Stderr, msg)

		printCount := 0

		for key, val := range mp.Tags {
			fmt.Fprintln(os.Stderr, "  - ["+fmt.Sprint(key)+"] "+val)
			printCount++

			if printCount >= 5 {
				if len(mp.Tags) > 5 {
					fmt.Fprintln(os.Stderr, "  ...")
				}
				break
			}
		}
	}

	return nil
}

func checkKeyForRecoveryCheckAlbums(vaultPath string, key []byte) error {
	filePath := path.Join(vaultPath, "albums.pmv")

	if !CheckFileExists(filePath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoAlbums",
				Other: "This vault does not have any albums.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		return nil
	}

	// Load file
	b, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	bJSON, err := encrypted_storage.DecryptFileContents(b, key)

	if err != nil {
		return err
	}

	var mp VaultAlbumsData

	err = json.Unmarshal(bJSON, &mp)

	if err != nil {
		return err
	}

	if len(mp.Albums) == 0 {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoAlbums",
				Other: "This vault does not have any albums.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
	} else {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultPrintAlbums",
				Other: "Vault albums (Total: {{.AlbumCount}}):",
			},
			TemplateData: map[string]interface{}{
				"AlbumCount": len(mp.Albums),
			},
		})
		fmt.Fprintln(os.Stderr, msg)

		printCount := 0

		for key, album := range mp.Albums {
			fmt.Fprintln(os.Stderr, "  - ["+fmt.Sprint(key)+"] "+album.Name)
			printCount++

			if printCount >= 5 {
				if len(mp.Albums) > 5 {
					fmt.Fprintln(os.Stderr, "  ...")
				}
				break
			}
		}
	}

	return nil
}

func readMainIndexFirstMedia(mainIndexPath string) (hasMedia bool, mediaId uint64, err error) {
	f, err := os.OpenFile(mainIndexPath, os.O_RDONLY, FILE_PERMISSION)

	if err != nil {
		return false, 0, err
	}

	_, err = f.Seek(0, 0)

	if err != nil {
		return false, 0, err
	}

	b := make([]byte, 8)

	_, err = f.Read(b)

	if err != nil {
		return false, 0, err
	}

	mediaCount := binary.BigEndian.Uint64(b)

	if mediaCount == 0 {
		return false, 0, nil
	}

	_, err = f.Seek(8, 0)

	if err != nil {
		return false, 0, err
	}

	b = make([]byte, 8)

	_, err = f.Read(b)

	if err != nil {
		return false, 0, err
	}

	mediaId = binary.BigEndian.Uint64(b)

	return true, mediaId, nil
}

func checkKeyForRecoveryCheckFirstMedia(vaultPath string, key []byte) error {
	mainIndexPath := path.Join(vaultPath, "main.index")

	if !CheckFileExists(mainIndexPath) {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoMedia",
				Other: "This vault does not have any media asset.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		return nil
	}

	hasMedia, mediaId, err := readMainIndexFirstMedia(mainIndexPath)

	if err != nil {
		return err
	}

	if !hasMedia {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultNoMedia",
				Other: "This vault does not have any media asset.",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		return nil
	}

	prefixByte := byte(mediaId % 256)
	prefixByteHex := hex.EncodeToString([]byte{prefixByte})

	mediaMetaFile := path.Join(vaultPath, "media", prefixByteHex, fmt.Sprint(mediaId), "meta.pmv")

	// Load file
	b, err := os.ReadFile(mediaMetaFile)

	if err != nil {
		return err
	}

	bJSON, err := encrypted_storage.DecryptFileContents(b, key)

	if err != nil {
		return err
	}

	var mp MediaMetadataMin

	err = json.Unmarshal(bJSON, &mp)

	if err != nil {
		return err
	}

	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultMediaSample",
			Other: "Sample media asset: [{{.MediaId}}] {{.MediaTitle}}",
		},
		TemplateData: map[string]interface{}{
			"MediaId":    fmt.Sprint(mp.Id),
			"MediaTitle": mp.Title,
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	return nil
}

func checkKeyForRecovery(vaultPath string, key []byte) {
	msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "VaultCheckingKey",
			Other: "Checking provided vault key...",
		},
	})
	fmt.Fprintln(os.Stderr, msg)

	// First, if tag file exists, decrypt it and print the first 5 tags

	err := checkKeyForRecoveryCheckTags(vaultPath, key)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultRecoverKeyProbablyInvalid",
				Other: "This error means the provided vault key is not valid or the vault is corrupted",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Second, if albums file exists, decrypt it and print the first 5 albums

	err = checkKeyForRecoveryCheckAlbums(vaultPath, key)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultRecoverKeyProbablyInvalid",
				Other: "This error means the provided vault key is not valid or the vault is corrupted",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// Finally, decrypt the fist media file

	err = checkKeyForRecoveryCheckFirstMedia(vaultPath, key)

	if err != nil {
		msg, _ := Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Error",
				Other: "Error: {{.Message}}",
			},
			TemplateData: map[string]interface{}{
				"Message": err.Error(),
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		msg, _ = Localizer.Localize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "VaultRecoverKeyProbablyInvalid",
				Other: "This error means the provided vault key is not valid or the vault is corrupted",
			},
		})
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}
