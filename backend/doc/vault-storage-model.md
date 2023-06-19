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

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `2` | Algorithm ID | Identifier of the algorithm, stored as a **Big Endian unsigned integer** |
| `2` | `H` | Header | Header containing any parameters required by the encryption algorithm. The size depends on the algorithm used. |
| `2 + H` | `N` | Body | Body containing the raw encrypted data.  The size depends on the initial unencrypted data and algorithm used.

The system is flexible enough to allow multiple encryption algorithms. Currently, there are 2 supported ones:

 - `AES256_FLAT`: ID = `1`, Uses ZLIB ([RFC 1950](https://datatracker.ietf.org/doc/html/rfc1950)) to compress the data, and then uses AES with a key of 256 bits to encrypt the data, CBC as the mode of operation and an IV of 128 bits. This algorithm uses a header containing the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `2 + H` | `4` | Compressed plaintext size | Size of the compressed plaintext, in bytes, used to remove padding |
| `2 + H + 4` | `16` | IV | Initialization vector for AES_256_CBC algorithm |

 - `AES256_FLAT`: ID = `2`, Uses AES with a key of 256 bits to encrypt the data, CBC as the mode of operation and an IV of 128 bits. This algorithm uses a header containing the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `2 + H` | `4` | Plaintext size | Size of the plaintext, in bytes, used to remove padding |
| `2 + H + 4` | `16` | IV | Initialization vector for AES_256_CBC algorithm |

### Index files

Index files have the `.index` extension.

They are sorted lists of media assets identifiers. They can store all the existing identifiers, or a fraction of them, for example, for a tag.

Thanks to being sorted, searching for a specific identifier can be achieved using binary search.

They are binary files, consisting of the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `8` | Index size | Number of entries the index file contains, stored as a **Big Endian unsigned integer** |
| `8 + 8*K` | `8` | Media asset identifier | Each media asset identifier is stored as a **Big Endian unsigned integer**. They are stored next to each other, and already sorted **from lower value to grater value** |

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

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `8` | File size | Size of the original file, in bytes, stored as a **Big Endian unsigned integer** |
| `8` | `8` | Chunk size limit | Max size of a chunk, in bytes, stored as a **Big Endian unsigned integer** |

After the header, the chunk index is stored. **For each chunk** the file was split into, the chunk index will store a metadata entry, withe the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `8` | Chunk pointer | Starting byte of the chunk, stored as a  **Big Endian unsigned integer** |
| `8` | `8` | Chunk size | Size of the chunk, in bytes, stored as a **Big Endian unsigned integer** |

After the chunk index, the encrypted chunks are stored following the same structure described in the [Encrypted JSON files](#encrypted-json-files) section.

This chunked structure allows to randomly access any point in the file as a low cost, since you don't need to decrypt the entire file, only the corresponding chunks. This capability is specially great for video rewinding and seeking.

#### Multi-File encrypted assets

These asset files are used to store multiple smaller files, meant to be sorted and accessed by an index number.

They are binary files consisting of 3 contiguous sections: The header, the file table and the encrypted files.

The header contains the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `8` | File count | Number of files stored by the asset, stored as a **Big Endian unsigned integer** |

After the header, a file table is stored. **For each file** stored by the asset, a metadata entry is stored, with the following fields:

| Starting byte | Size (bytes) | Value name | Description |
|---|---|---|---|
| `0` | `8` | File data pointer | Starting byte of the file encrypted data, stored as a  **Big Endian unsigned integer** |
| `8` | `8` | File size | Size of the encrypted file, in bytes, stored as a **Big Endian unsigned integer** |

After the file table, each file is stored following the same structure described in the [Encrypted JSON files](#encrypted-json-files) section.

This format is useful to store video previews, without the need to use too many files.

## Vault folder structure

Media vaults are stored in folders. A vault folder may contain the following files and folders:

| Name | Path | Type | Description |
| --- | --- | --- | --- |
| **Media assets** | `media` | Folder | Folder where media assets are stored. |
| **Tag indexes** | `tags` | Folder | Folder where tag indexes are stored. |
| **Lock file** | `vault.lock` | [Lock file](#lock-file) | File used to prevent multiple instances of the PersonalMediaVault backend to access a vault at the same time. It may not be present, in case the vault is not being accessed. |
| **Launcher configuration** | `launcher.config.json` | [Unencrypted JSON file](#unencrypted-json-files) | File to store the launcher configuration. |
| **Hashed Credentials** | `credentials.json` | [Unencrypted JSON file](#unencrypted-json-files) | File to store the existing accounts, along with the hashed credentials and the encrypted vault key, protected with the account password. |
| **Media ID tracker** | `media_ids.json` | [Unencrypted JSON file](#unencrypted-json-files) | File to store the last used media asset ID. |
| **Tasks tracker** | `tasks.json` | [Unencrypted JSON file](#unencrypted-json-files) | File used to store the last used task ID, along with the list of pending tasks. |
| **Albums** | `albums.pmv` | [Encrypted JSON file](#encrypted-json-files) | File used to store the existing albums, including the metadata and the list of media assets included in them. |
| **Tag list** | `tag_list.pmv` | [Encrypted JSON file](#encrypted-json-files) | File to store the metadata of the existing vault tags |
| **User configuration** | `user_config.pmv` | [Encrypted JSON file](#encrypted-json-files) | File to store user configuration, like the vault title or the encoding parameters |
| **Main index** | `main.index` | [Index file](#index-files) | File to index every single media asset existing in the vault. |

### Media assets folder


