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
| TEMP_PATH | Temporal path to store things like uploaded files or to use for FFMPEG encoding.  Note: It should be in a different filesystem if the vault is stored in an unsafe environment. By default, this will be stored in `~/.pmv/temp` |
| FRONTEND_PATH | Path to static frontend to serve it. |
| SSL_CERT | Path to the SSL certificate. Required to enable HTTPS |
| SSL_KEY | Path to SSL private key. Required to enable HTTPS |
| USING_PROXY | Set it to `YES` if you are using a reverse proxy. |
| TEMP_FILE_DELETE_MODE | Set it to `SECURE` to clear all the bytes of temporal files (secure delete). |

## Documentation

 - [Vault Storage Model](./doc/vault-storage-model.md)
 - [API documentation](./doc/api.md)
 - [Frontend customization](./doc/frontend-customization.md)
