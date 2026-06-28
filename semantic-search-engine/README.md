# Semantic Search Engine

The semantic search is responsible for the following features for PersonalMediaVault:

- Semantic search of images / titles: Instead of searching with tags or exact words in the title, semantic search allows the user to perform search in natural language related to the meaning of the title or the image. Example: If the user searches `Black Cat`, images of blacks cats in their vault will appear first.
- Search by image similarity: The user selects an image, and similar images will appear first.

In order to achieve this functionality, a multi-modal embedding model is needed, which supports both images and text. This embedding model will turn both images and text into vectors of the same size. Then, by sorting vectors by distance, semantic search is achieved.

The model type chosen for this task is [OpenCLIP](https://github.com/mlfoundations/open_clip), an open source version of the original CLIP model, developed by OpenAI.

The model inference is achieved using the [ONNX Runtime](https://github.com/microsoft/onnxruntime), so models need to be adapted for this runtime.

## Internal server

The inference runs in an internal Rust server that exposes a GRPC service for both text and image embeddings.

The source code for this server can be found in the [server](./server/) folder.

## Database

The internal server runs an [SQLite](https://sqlite.org/) database, with the following extensions:

- [sqlite-vec](https://github.com/asg017/sqlite-vec) - Extension to store vectors in the database and query by distance.
- [sqlcipher](https://github.com/sqlcipher/sqlcipher) - Extension to encrypt the database, in order to protect it the same way the rest of the vault is protected.

The database is stored inside the vault folder, with name `semantic-search.db`.

The database is encrypted with a passphrase generated from the vault key:

- Concatenate the vault key with the UTF-8 bytes of the string `semantic-search-db-key`.
- Take the SHA-256 hash of the concatenation result.
- Encode this hash into hexadecimal with lowercase characters.

```js
passphrase = hex_lowercase(
  sha_256(concatenate_bytes(vault_key, utf8_bytes("semantic-search-db-key"))),
);
```

### Database tables

The SQLite database contains 2 tables:

- `config` - A generic key-value table to store configuration or internal state.
- `vectors` - The table where the vectors are stored.

Fields for the `config` table:

| Name    | Type   | Key Type      | Description                                    |
| ------- | ------ | ------------- | ---------------------------------------------- |
| `key`   | `TEXT` | `PRIMARY KEY` | The key for the configuration setting or state |
| `value` | `TEXT` | -             | The value of the setting or state.             |

Fields for the `vectors` table:

| Name          | Type                       | Key Type      | Description                                               |
| ------------- | -------------------------- | ------------- | --------------------------------------------------------- |
| `id`          | `INTEGER`                  | `PRIMARY KEY` | Unique identifier of the vector.                          |
| `media_id`    | `INTEGER`                  | -             | ID of the media resource the vector was created for.      |
| `vector_type` | `INTEGER`                  | -             | Vector type. `0` for text vectors, `1` for image vectors. |
| `data_hash`   | `TEXT`                     | -             | Hash of the data to check by synchronization processes.   |
| `embedding`   | `float[{MODEL_DIMENSION}]` | -             | The vector. Same dimensions as the embedding model.       |

## Protocol

The communication protocol between the PersonalMediaVault backend and the server is GRPC, using protocol buffers.

You can find the protocol files in the [protocol](./protocol/) folder.
