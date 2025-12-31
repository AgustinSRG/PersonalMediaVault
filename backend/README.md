# Personal Media Vault (Backend)

Personal media vault backend component, implemented in golang. It stores the media assets, encodes them and provides an HTTP API for clients to access the vault.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/PersonalMediaVault/backend
```

To compile the code type:

```
go build -o pmvd
```

The build command will create a binary in the current directory, called `pmvd`, or `pmvd.exe` if you are using Windows.

## Usage

In order to run the project, you can run the `pmvd` binary.

In order to see the options, use:

```
pmvd --help
```

In order to run the daemon, use:

```
pmvd --daemon [OPTIONS]
```

Here is the full list of available options:

| Option | Description |
|---|---|
| `--help, -h` | Prints command line options. |
| `--version, -v` | Prints version. |
| `--daemon, -d` | Runs backend daemon. |
| `--init, -i` | Initializes the vault. Asks for username and password. |
| `--clean, -c` | Cleans temporal path before starting the daemon. |
| `--port -p <port>` | Sets the listening port. By default 80 (or 443 if using SSL). |
| `--bind -b <bind-addr>` | Sets the bind address. By default it binds all interfaces. |
| `--vault-path, -vp <path>` | Sets the data storage path for the vault. |
| `--cache-size <size>` | Sets the LRU cache size. By default is can hold 1024 elements. |

Also, here is a list of available debug options:

| Option | Description |
|---|---|
| `--skip-lock` | Ignores vault lockfile. |
| `--fix-consistency` | Fixes vault consistency at startup (takes some time). |
| `--debug` | Enables debug mode. |
| `--log-requests` | Enables logging requests to standard output. |
| `--cors-insecure` | Allows all CORS requests (insecure, for development). |
| `--launch-tag <tag>` | Sets launcher tag (for launcher use). |

Also, here is a list of environment variables to configure other options:

| Variable Name | Description |
|---|---|
| FFMPEG_PATH | Path to `ffmpeg` binary |
| FFPROBE_PATH | Path to `ffprobe` binary |
| FFMPEG_VIDEO_ENCODER | Name of the FFmpeg codec to encode the videos. |
| TEMP_PATH | Temporal path to store things like uploaded files or to use for FFMPEG encoding.  Note: It should be in a different filesystem if the vault is stored in an unsafe environment. By default, this will be stored in `~/.pmv/temp` |
| FRONTEND_PATH | Path to static frontend to serve it. |
| SSL_CERT | Path to the SSL certificate. Required to enable HTTPS |
| SSL_KEY | Path to SSL private key. Required to enable HTTPS |
| USING_PROXY | Set it to `YES` if you are using a reverse proxy. |
| VAULT_INITIAL_USER | The initial vault username to set if the vault folder is empty. |
| VAULT_INITIAL_PASSWORD | The initial vault password to set if the vault folder is empty. |

Also, here is the list environment variables to configure semantic search:

| Variable Name | Description |
|---|---|
| SEMANTIC_SEARCH_ENABLED | Set it to `YES` to enable semantic search. The rest of the options must be configured. Otherwise you will get an error. |
| QDRANT_HOST | Host of the Qdrant database. |
| QDRANT_PORT | GRPC port of the Qdrant database. Default: `6334` |
| QDRANT_API_KEY | API key for the Qdrant database. |
| QDRANT_USE_TLS | Set it to `YES` in order to use TLS to connect to the Qdrant database. |
| QDRANT_INITIAL_SCAN | By default, when the vault is unlocked, a task to scan for missing media in the Qdrant database will be created. Set this variable to `NO` to disable it. |
| CLIP_API_BASE | Base URL of the CLIP API (provided by `pmv-ai-service`). Example: `http://localhost:5000/clip` |
| CLIP_API_AUTH | Value of the `Authorization` header in order to use the CLIP API. |
| CLIP_IMAGE_SIZE_LIMIT_MB | Limit on size (MB) before the images are discarded from being encoded by CLIP. Default: `20` |

## Documentation

 - [Vault Storage Model](./doc/vault-storage-model.md)
 - [API documentation](./doc/api.md)
 - [Frontend customization](./doc/frontend-customization.md)
