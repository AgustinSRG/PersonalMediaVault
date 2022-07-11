# Personal Media Vault (Backend)

Personal media vault backend component, implemented in golang. It stores the media assets, encodes them and provides an HTTP API for clients to access the vault.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/PersonalMediaVault/backend
```

To compile the code type:

```
go build -o pmv
```

The build command will create a binary in the currenct directory, called `pmv`, or `pmv.exe` if you are using Windows.

## Usage

In order to run the project, you can run the `pmv` binary.

In order to see the options, use:

```
pmv --help
```

In order to run the daemon, use:

```
pmv --daemon [OPTIONS]
```
