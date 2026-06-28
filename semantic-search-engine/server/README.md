# Semantic search engine server

This server is responsible for running the sematic search capabilities of PersonalMediaVault:

 - It loads an OpenClip model and exposes encoding methods to the backend.
 - It stores vectors into a SQLite database and exposes methods to insert, delete and fix vectors.

The GRPC server will bind on a random free port and the chosen port will be printed to the standard output.
The backend can then read this port and communicate as the GRPC client with this server.

## Command line usage

**Usage:**

```
pmv-sse [OPTIONS] <MODEL-PATH> <SQLITE-DB-PATH>
```

**Options:**

| Option                  | Description                                                          |
| ----------------------- | -------------------------------------------------------------------- |
| `-l --loglevel <LEVEL>` | Sets the log level. It can be: `ERROR`, `WARNING`, `INFO` or `DEBUG` |

The **Model path** must be the path to the folder containing the ONXX OpenCLIP compatible model. More information: https://github.com/RuurdBijlsma/open-clip-inference-rs#model-support

The **SQLite Database Path** is the path to a sqlite database file where all the vectors will be stored.

## Environment variables

| Environment variable       | Description                                                  |
| -------------------------- | ------------------------------------------------------------ |
| `API_KEY`                  | A secret API key to restrict access to the internal service. |
| `SQLITE_CIPHER_PASSPHRASE` | An encryption passphrase for the sqlite database.            |

## Building

Type the following command to compile:

```sh
cargo build --release
```

The resulting binaries will be placed in the `target` folder.
