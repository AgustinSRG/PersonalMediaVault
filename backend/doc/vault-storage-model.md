# Vault storage model

This is the documentation for the vault storage model, including the types of files it uses, their internal structure and the encryption algorithms used.

Use this document as reference for any software development that requires interaction with the vault files.

## File types

The vault storage model uses different types of files:

 - [**Lock file**](#lock-file): File used to prevent multiple instances of PersonalMediaVault accessing the same vault.
 - [**Unencrypted JSON files**](#unencrypted-json-files): Configuration files that do not contain any protected vault data.
 - [**Encrypted JSON files**](#encrypted-json-files): Used to store metadata.
 - [**Index files**](#index-files): Used to store lists of media asset IDs, in order to make searching faster.
 - [**Encrypted assets**](#encrypted-assets): Encrypted files containing the media assets. They can be single-file or multi-file.

### Lock file

The lock file has the `.lock` extension.

It stores in plain text, a decimal number representing the PID of the current Process accessing the vault.

PersonalMediaVault backend should check for the existence of this file and the process before accessing the vault.

### Unencrypted JSON files

Unencrypted JSON files have the `.json` extension.

They follow the [JSON format](https://www.json.org/). The schema varies depending on the specific file.

Since they are not encrypted, they just store configuration, like the port it should listen, or the encryption parameters.

### Encrypted JSON files

Encrypted JSON files have the `.pmv` extension.

They take as a base a JSON plaintext, that is encrypted using an algorithm like AES.

They are binary files, with the following structure:

| Starting byte | Size (bytes) | Value name   | Description                                                                                                    |
| ------------- | ------------ | ------------ | -------------------------------------------------------------------------------------------------------------- |
| `0`           | `2`          | Algorithm ID | Identifier of the algorithm, stored as a **Big Endian unsigned integer**                                       |
| `2`           | `H`          | Header       | Header containing any parameters required by the encryption algorithm. The size depends on the algorithm used. |
| `2 + H`       | `N`          | Body         | Body containing the raw encrypted data.  The size depends on the initial unencrypted data and algorithm used.  |

The system is flexible enough to allow multiple encryption algorithms. Currently, there are 2 supported ones:

 - `AES256_FLAT`: ID = `1`, Uses ZLIB ([RFC 1950](https://datatracker.ietf.org/doc/html/rfc1950)) to compress the data, and then uses AES with a key of 256 bits to encrypt the data, CBC as the mode of operation and an IV of 128 bits. This algorithm uses a header containing the following fields:

| Starting byte | Size (bytes) | Value name                | Description                                                        |
| ------------- | ------------ | ------------------------- | ------------------------------------------------------------------ |
| `2 + H`       | `4`          | Compressed plaintext size | Size of the compressed plaintext, in bytes, used to remove padding |
| `2 + H + 4`   | `16`         | IV                        | Initialization vector for AES_256_CBC algorithm                    |

 - `AES256_FLAT`: ID = `2`, Uses AES with a key of 256 bits to encrypt the data, CBC as the mode of operation and an IV of 128 bits. This algorithm uses a header containing the following fields:

| Starting byte | Size (bytes) | Value name     | Description                                             |
| ------------- | ------------ | -------------- | ------------------------------------------------------- |
| `2 + H`       | `4`          | Plaintext size | Size of the plaintext, in bytes, used to remove padding |
| `2 + H + 4`   | `16`         | IV             | Initialization vector for AES_256_CBC algorithm         |

### Index files

Index files have the `.index` extension.

They are sorted lists of media assets identifiers. They can store all the existing identifiers, or a fraction of them, for example, for a tag.

Thanks to being sorted, searching for a specific identifier can be achieved using binary search.

They are binary files, consisting of the following fields:

| Starting byte | Size (bytes) | Value name             | Description                                                                                                                                                             |
| ------------- | ------------ | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `0`           | `8`          | Index size             | Number of entries the index file contains, stored as a **Big Endian unsigned integer**                                                                                  |
| `8 + 8*K`     | `8`          | Media asset identifier | Each media asset identifier is stored as a **Big Endian unsigned integer**. They are stored next to each other, and already sorted **from lower value to grater value** |

### Encrypted assets

Encrypted assets have the `.pma` extension.

They stored one or multiple encrypted files.

They are also binary files, and they can be of two types:

 - [**Single-File encrypted assets**](#single-file-encrypted-assets): They store a single file in size-limited chunks. Their name usually starts with `s_`.
 - [**Multi-File encrypted assets**](#multi-file-encrypted-assets): They store multiple smaller files. Their name usually starts with `m_`.

#### Single-File encrypted assets

These asset files are used to store a single and possibly big file in chunks, encrypted each chunk using the same method described by the [Encrypted JSON files](#encrypted-json-files) section.

They are binary files consisting of 3 contiguous sections: The header, the chunk index and the encrypted chunks.

The header contains the following fields:

| Starting byte | Size (bytes) | Value name       | Description                                                                      |
| ------------- | ------------ | ---------------- | -------------------------------------------------------------------------------- |
| `0`           | `8`          | File size        | Size of the original file, in bytes, stored as a **Big Endian unsigned integer** |
| `8`           | `8`          | Chunk size limit | Max size of a chunk, in bytes, stored as a **Big Endian unsigned integer**       |

After the header, the chunk index is stored. **For each chunk** the file was split into, the chunk index will store a metadata entry, withe the following fields:

| Starting byte | Size (bytes) | Value name    | Description                                                              |
| ------------- | ------------ | ------------- | ------------------------------------------------------------------------ |
| `0`           | `8`          | Chunk pointer | Starting byte of the chunk, stored as a  **Big Endian unsigned integer** |
| `8`           | `8`          | Chunk size    | Size of the chunk, in bytes, stored as a **Big Endian unsigned integer** |

After the chunk index, the encrypted chunks are stored following the same structure described in the [Encrypted JSON files](#encrypted-json-files) section.

This chunked structure allows to randomly access any point in the file as a low cost, since you don't need to decrypt the entire file, only the corresponding chunks. This capability is specially great for video rewinding and seeking.

#### Multi-File encrypted assets

These asset files are used to store multiple smaller files, meant to be sorted and accessed by an index number.

They are binary files consisting of 3 contiguous sections: The header, the file table and the encrypted files.

The header contains the following fields:

| Starting byte | Size (bytes) | Value name | Description                                                                      |
| ------------- | ------------ | ---------- | -------------------------------------------------------------------------------- |
| `0`           | `8`          | File count | Number of files stored by the asset, stored as a **Big Endian unsigned integer** |

After the header, a file table is stored. **For each file** stored by the asset, a metadata entry is stored, with the following fields:

| Starting byte | Size (bytes) | Value name        | Description                                                                            |
| ------------- | ------------ | ----------------- | -------------------------------------------------------------------------------------- |
| `0`           | `8`          | File data pointer | Starting byte of the file encrypted data, stored as a  **Big Endian unsigned integer** |
| `8`           | `8`          | File size         | Size of the encrypted file, in bytes, stored as a **Big Endian unsigned integer**      |

After the file table, each file is stored following the same structure described in the [Encrypted JSON files](#encrypted-json-files) section.

This format is useful to store video previews, without the need to use too many files.

## Vault folder structure

Media vaults are stored in folders. A vault folder may contain the following files and folders:

| Name                                               | Path               | Type                                             | Description                                                                                                                                                                   |
| -------------------------------------------------- | ------------------ | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [**Media assets**](#media-assets-folder)           | `media`            | Folder                                           | Folder where media assets are stored.                                                                                                                                         |
| [**Tag indexes**](#tag-indexes-folder)             | `tags`             | Folder                                           | Folder where tag indexes are stored.                                                                                                                                          |
| [**Lock file**](#lock-file)                        | `vault.lock`       | [Lock file](#lock-file)                          | File used to prevent multiple instances of the PersonalMediaVault backend to access a vault at the same time. It may not be present, in case the vault is not being accessed. |
| [**Credentials file**](#credentials-file)          | `credentials.json` | [Unencrypted JSON file](#unencrypted-json-files) | File to store the existing accounts, along with the hashed credentials and the encrypted vault key, protected with the account password.                                      |
| [**Media ID tracker**](#media-id-tracker)          | `media_ids.json`   | [Unencrypted JSON file](#unencrypted-json-files) | File to store the last used media asset ID.                                                                                                                                   |
| [**Tasks tracker**](#tasks-tracker)                | `tasks.json`       | [Unencrypted JSON file](#unencrypted-json-files) | File used to store the last used task ID, along with the list of pending tasks.                                                                                               |
| [**Albums**](#albums-file)                         | `albums.pmv`       | [Encrypted JSON file](#encrypted-json-files)     | File used to store the existing albums, including the metadata and the list of media assets included in them.                                                                 |
| [**Tag list**](#tags-file)                         | `tag_list.pmv`     | [Encrypted JSON file](#encrypted-json-files)     | File to store the metadata of the existing vault tags                                                                                                                         |
| [**User configuration**](#user-configuration-file) | `user_config.pmv`  | [Encrypted JSON file](#encrypted-json-files)     | File to store user configuration, like the vault title or the encoding parameters                                                                                             |
| [**Main index**](#main-index-file)                 | `main.index`       | [Index file](#index-files)                       | File to index every single media asset existing in the vault.                                                                                                                 |

### Media assets folder

The media assets are stored inside the `media` folder.

In order to prevent the folder size to increase too much, the assets are distributed evenly in **256** sub-folders. The sub-folder name for each media asset is calculated from its identifier, since it's a 64 bit unsigned integer, the folder name is the **identifier module 256**, and the result turned into a **2 character hex lowercased string**

Examples: `00`, `01`, `02`..., `fd`, `fe`, `ff`.

Inside each subfolder, the assets are stored inside their own folders, named by turning their identifier into a decimal string. Examples:

 - `media_id=0` - Stored inside `{VAULT_FOLDER}/media/00/0`
 - `media_id=15` - Stored inside `{VAULT_FOLDER}/media/0f/15`

```go
import (
    "fmt",
    "hex", 
    "path",
)

func GetMediaAssetFolder(vault_path string, media_id uint64) string {
    subFolderName := hex.EncodeToString([]byte{ byte(media_id % 256) });

    return path.Join(vault_path, "media", subFolderName, fmt.Sprint(media_id))
}
```

The media asset folder may contain up to 3 types of files:

 - [Media asset metadata file](#media-asset-metadata-file): Named `meta.pmv` and used to store metadata.
 - [Single-File assets](#single-file-encrypted-assets): Named concatenating the `s_` prefix and the asset ID in decimal, with `.pma` extension.
 - [Multi-File assets](#multi-file-encrypted-assets): Named concatenating the `m_` prefix and the asset ID in decimal, with `.pma` extension.

#### Media asset metadata file

Each media asset folder must contain a file named `meta.pmv`, being an [encrypted JSON file](#encrypted-json-files) containing the metadata of the media asset.

The file contains the following fields:

| Field name              | Type                                          | Description                                                                                                                   |
| ----------------------- | --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `id`                    | Number (64 bit unsigned integer)              | Media asset identifier                                                                                                        |
| `type`                  | Number (8 bit unsigned integer)               | Media type. Can be: `1` (Image), `2` (Video / Animation) or `3` (Audio / Sound)                                               |
| `title`                 | String                                        | Title                                                                                                                         |
| `description`           | String                                        | Description                                                                                                                   |
| `tags`                  | Array&lt;Number (64 bit unsigned integer)&gt; | List of tags for the media. Only identifiers are stored                                                                       |
| `duration`              | Number (Floating point)                       | Duration of the media in seconds                                                                                              |
| `width`                 | Number (32 bit unsigned integer)              | Width in pixels                                                                                                               |
| `height`                | Number (32 bit unsigned integer)              | Height in pixels                                                                                                              |
| `fps`                   | Number (32 bit unsigned integer)              | Frames per second                                                                                                             |
| `upload_time`           | Number (64 bit integer)                       | Upload timestamp (Unix milliseconds format)                                                                                   |
| `next_asset_id`         | Number (64 bit unsigned integer)              | Identifier to use for the next asset, when created                                                                            |
| `original_ready`        | Boolean                                       | True if the original asset exists and is ready                                                                                |
| `original_asset`        | Number (64 bit unsigned integer)              | Asset ID of the original asset. The original asset is Single-File                                                             |
| `original_ext`          | String                                        | Extension of the original asset file. Eg: `mp4`                                                                               |
| `original_encoded`      | Boolean                                       | True if the original asset is encoded                                                                                         |
| `original_task`         | Number (64 bit unsigned integer)              | If the original asset is not encoded, the ID of the task assigned to encode it                                                |
| `thumb_ready`           | Boolean                                       | True if the thumbnail asset exists and is ready                                                                               |
| `thumb_asset`           | Number (64 bit unsigned integer)              | Asset ID of the thumbnail asset. The thumbnail asset is Single-File                                                           |
| `previews_ready`        | Boolean                                       | True if the video previews asset exists and is ready                                                                          |
| `previews_asset`        | Number (64 bit unsigned integer)              | Asset ID of the video previews asset. The video previews asset is Multi-File                                                  |
| `previews_interval`     | Number (Floating point)                       | Video previews interval in seconds                                                                                            |
| `previews_task`         | Number (64 bit unsigned integer)              | If the video previews asset is not ready, the ID of the task assigned to generate it                                          |
| `force_start_beginning` | Boolean                                       | True to indicate the player not to store the current playing time, so the video or audio starts from the beginning every time |
| `img_notes`             | Boolean                                       | True if the image has a notes asset                                                                                           |
| `img_notes_asset`       | Number (64 bit unsigned integer)              | Asset ID of the image notes asset. The image notes asset is Single-File                                                       |
| `resolutions`           | Array&lt;Resolution&gt;                       | List of extra resolutions                                                                                                     |
| `subtitles`             | Array&lt;Subtitle&gt;                         | List of subtitles files                                                                                                       |
| `time_splits`           | Array&lt;TimeSplit&gt;                        | List of time splits for videor or audios                                                                                      |
| `audio_tracks`          | Array&lt;AudioTrack&gt;                       | List of extra audio tracks for videos                                                                                         |

The `Resolution` object has the following fields:

| Field name | Type                             | Description                                                     |
| ---------- | -------------------------------- | --------------------------------------------------------------- |
| `width`    | Number (32 bit unsigned integer) | Width in pixels                                                 |
| `height`   | Number (32 bit unsigned integer) | Height in pixels                                                |
| `fps`      | Number (32 bit unsigned integer) | Frames per second                                               |
| `ready`    | Boolean                          | True if the asset is ready                                      |
| `asset`    | Number (64 bit unsigned integer) | Asset ID of the asset. The asset is Single-File                 |
| `ext`      | String                           | Asset file extension. Example: `mp4`                            |
| `task_id`  | Number (64 bit unsigned integer) | If the asset is not ready, ID of the task assigned to encode it |
 
The `Subtitle` object has the following fields:

| Field name | Type                             | Description                                     |
| ---------- | -------------------------------- | ----------------------------------------------- |
| `id`       | String                           | Subtitles language identifier. Example: `eng`   |
| `name`     | String                           | Subtitles file name. Example `English`          |
| `asset`    | Number (64 bit unsigned integer) | Asset ID of the asset. The asset is Single-File |

The `TimeSplit` object has the following fields:

| Field name | Type                    | Description                            |
| ---------- | ----------------------- | -------------------------------------- |
| `time`     | Number (Floating point) | Time in seconds where the split starts |
| `name`     | String                  | Name of the time split                 |

The `AudioTrack` object has the following fields:

| Field name | Type                             | Description                                     |
| ---------- | -------------------------------- | ----------------------------------------------- |
| `id`       | String                           | Audio track language identifier. Example: `eng` |
| `name`     | String                           | Audio track file name. Example `English`        |
| `asset`    | Number (64 bit unsigned integer) | Asset ID of the asset. The asset is Single-File |

The image notes asset is a JSON file, containing an array of `ImageNote` objects, with the following fields:

| Field name | Type                    | Description                            |
| ---------- | ----------------------- | -------------------------------------- |
| `x`        | Number (32 bit integer) | X position (pixels)                    |
| `y`        | Number (32 bit integer) | Y position (pixels)                    |
| `w`        | Number (32 bit integer) | Width (pixels)                         |
| `h`        | Number (32 bit integer) | Height (pixels)                        |
| `text`     | String                  | Text to display for the specified area |

### Tag indexes folder

When a tag is added to the vault, a new [index file](#index-files) is created inside the `tags` folder, with a name made by concatenating the `tag_` prefix with the tag identifier encoded in decimal, and the `.index` extension.

```go
import (
    "fmt",
    "path",
)

func GetTagIndexPath(vault_path string, tag_id uint64) string {
	return path.Join(vault_path, "tags", "tag_"+fmt.Sprint(tag_id)+".index")
}
```

Each tag index file contains the list of media asset identifiers that have such tag.
### Credentials file

The credentials file, named `credentials.json` is an [unencrypted JSON file](#unencrypted-json-files) used to store the hashed credentials, along with the encrypted vault key.

The JSON file contains the following fields:

| Field name    | Type                 | Description                                  |
| ------------- | -------------------- | -------------------------------------------- |
| `user`        | String               | Username of the root account                 |
| `pwhash`      | String               | Password hash. Base 64 encoded               |
| `salt`        | String               | Hashing salt. Base 64 encoded                |
| `enckey`      | String               | Encrypted key. Base 64 encoded               |
| `method`      | String               | Name of the hashing + encryption method used |
| `fingerprint` | String               | Vault fingerprint                            |
| `accounts`    | Array&lt;Account&gt; | Array of additional accounts                 |

Each `Account` is an object with the following fields:

| Field name | Type    | Description                                            |
| ---------- | ------- | ------------------------------------------------------ |
| `user`     | String  | Account username                                       |
| `pwhash`   | String  | Password hash. Base 64 encoded                         |
| `salt`     | String  | Hashing salt. Base 64 encoded                          |
| `enckey`   | String  | Encrypted key. Base 64 encoded                         |
| `method`   | String  | Name of the hashing + encryption method used           |
| `write`    | Boolean | True if the account has permission to modify the vault |

Currently, the following methods are implemented:

 - [AES256 + SHA256 + SALT16](#aes256--sha256--salt16) - Identifier: `aes256/sha256/salt16`

#### AES256 + SHA256 + SALT16

This algorithm uses a **random salt of 16 bytes** (128 bits).

The password hash is calculated by using the SHA256 algorithm 2 times on the binary concatenation of the password (as UTF-8) and the random salt:

```go
import (
    "sha256"
)

func ComputePasswordHash(password string, salt []byte) []byte {
	firstHash := sha256.Sum256(append([]byte(password), salt...))
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]
}
```

The vault ket is encrypted using the AES256 algorithm, using the system defined in the [Encrypted JSON files](#encrypted-json-files) section. Specifically using the `AES256_FLAT` mode.

The key for the encryption is calculated by hashing with SHA256 the the binary concatenation of the password (as UTF-8) and the random salt:

```go
import (
    "sha256"
)

func ComputeAESEncryptionKey(password string, salt []byte) []byte {
	passwordHash := sha256.Sum256(append([]byte(password), salt...))
	return passwordHash[:]
}
```

### Media ID tracker

The media ID tracker file, named `media_ids.json` is an [unencrypted JSON file](#unencrypted-json-files) used to store the number of used media identifiers, very important to prevent duplicated identifiers.

The JSON file has just one field:

| Field name | Type                             | Description                                          |
| ---------- | -------------------------------- | ---------------------------------------------------- |
| `next_id`  | Number (64 bit unsigned integer) | Next identifier to use when adding a new media asset |

### Tasks tracker

The task tracker file, named `tasks.json`  is an [unencrypted JSON file](#unencrypted-json-files) used to store the number of used task identifiers, in order to prevent duplicates. It also stores the pending tasks, in order to continue them in case of a vault restart.

The JSON file contains the following fields:

| Field name | Type                                      | Description                                                           |
| ---------- | ----------------------------------------- | --------------------------------------------------------------------- |
| `next_id`  | Number (64 bit unsigned integer)          | Next identifier to use when creating a new task                       |
| `pending`  | Object (Mapping String -&gt; PendingTask) | Mapping. For each pending task, the required metadata to restart them |

The `PendingTask` objects have the following fields:

| Field name       | Type                                                                            | Description                                                                                                 |
| ---------------- | ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `id`             | Number (64 bit unsigned integer)                                                | Task identifier                                                                                             |
| `media_id`       | Number (64 bit unsigned integer)                                                | Media asset ID                                                                                              |
| `type`           | Number (8 bit unsigned integer)                                                 | Task type. It can be: `0` (Encode original), `1` (Encode extra resolution) or `2` (Generate video previews) |
| `first_time_enc` | Boolean                                                                         | True if this task is the first time the asset is being encoded (was just uploaded)                          |
| `resolution`     | Object { `width`: Width (px), `height`: Height (px), `fps`: Frames per second } | Resolution for type = `1`                                                                                   |

### Albums file

The albums file, named `albums.pmv` is an [encrypted JSON file](#encrypted-json-files) used to store the list of existing albums in the vault.

The file has the following fields:

| Field name | Type                              | Description                                                           |
| ---------- | --------------------------------- | --------------------------------------------------------------------- |
| `next_id`  | Number (64 bit unsigned integer)  | Identifier to use for the next album, when creating a new one.        |
| `albums`   | Object { Mapping ID -&gt; Album } | List of albums. For each album it maps its identifier to its metadata |

The `Album` object has the following fields:

| Field name | Type                                          | Description                                            |
| ---------- | --------------------------------------------- | ------------------------------------------------------ |
| `name`     | String                                        | Name of the album                                      |
| `lm`       | Number (64 bit integer)                       | Last modified timestamp. Unix milliseconds format      |
| `list`     | Array&lt;Number (64 bit unsigned integer)&gt; | List of media asset identifiers contained in the album |

### Tags file

The tags file, named `tag_list.pmv` is an [encrypted JSON file](#encrypted-json-files) used to store the list of existing tags in the vault.

The file has the following fields:

| Field name | Type                               | Description                                                    |
| ---------- | ---------------------------------- | -------------------------------------------------------------- |
| `next_id`  | Number (64 bit unsigned integer)   | Identifier to use for the next tag, when creating a new one.   |
| `tags`     | Object { Mapping ID -&gt; String } | List of tags. For each tag, it maps its identifier to its name |

### User configuration file

The user configuration file, named `user_config.pmv` is an [encrypted JSON file](#encrypted-json-files) used to store the vault configuration set by the user.

The file has the following fields:

| Field name                | Type                         | Description                                                 |
| ------------------------- | ---------------------------- | ----------------------------------------------------------- |
| `title`                   | String                       | Vault custom title                                          |
| `css`                     | String                       | Custom CSS for the frontend                                 |
| `max_tasks`               | Number (32 bit integer)      | Max number of tasks to run in parallel                      |
| `encoding_threads`        | Number (32 bit integer)      | Max number of threads to use for a single encoding task     |
| `video_previews_interval` | Number (32 bit integer)      | Video previews interval (seconds)                           |
| `resolutions`             | Array&lt;VideoResolution&gt; | Resolutions to automatically encode when uploading a video  |
| `image_resolutions`       | Array&lt;ImageResolution&gt; | Resolutions to automatically encode when uploading an image |

The `VideoResolution` object has the following fields:

| Field name | Type                             | Description       |
| ---------- | -------------------------------- | ----------------- |
| `width`    | Number (32 bit unsigned integer) | Width in pixels   |
| `height`   | Number (32 bit unsigned integer) | Height in pixels  |
| `fps`      | Number (32 bit unsigned integer) | Frames per second |

The `ImageResolution` object has the following fields:

| Field name | Type                             | Description      |
| ---------- | -------------------------------- | ---------------- |
| `width`    | Number (32 bit unsigned integer) | Width in pixels  |
| `height`   | Number (32 bit unsigned integer) | Height in pixels |

### Main index file

The main index file, named `main.index` is an [index file](#index-files) containing every single media asset identifier existing in the vault.

This file is used to check if a media asset exists and to perform searches when a tag filter is not specified.
