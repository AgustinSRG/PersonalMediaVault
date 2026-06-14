# Semantic search engine server

This server is responsible for loading an OpenCLIP model and exposing a GRPC server to use it.

The GRPC server will bind on a random free port and the chosen port will be printed to the standard output.

## Command line usage

**Usage:**

```
pmv-sse [OPTIONS] <MODEL-PATH>
```

**Options:**

| Option                  | Description                                                          |
| ----------------------- | -------------------------------------------------------------------- |
| `-l --loglevel <LEVEL>` | Sets the log level. It can be: `ERROR`, `WARNING`, `INFO` or `DEBUG` |

The **Model path** must be the path to the folder containing the ONXX OpenCLIP compatible model. More information: https://github.com/RuurdBijlsma/open-clip-inference-rs#model-support

## Environment variables

| Environment variable | Description                                                  |
| -------------------- | ------------------------------------------------------------ |
| `API_KEY`            | A secret API key to restrict access to the internal service. |


## Building

Type the following command to compile:

```sh
cargo build --release
```

The resulting binaries will be placed in the `target` folder.
