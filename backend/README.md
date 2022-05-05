# Personal Media Vault (Backend)

Personal media vault backend component, implemented in golang. It stores the media assets, encodes them and provides an HTTP API for clients to access the vault.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/PersonalMediaVault/backend
```

To compile the code type:

```
go build -o 
```

The build command will create a binary in the currenct directory, called `personal-media-vault`, or `personal-media-vault.exe` if you are using Windows.

## Usage

In order to run the project, you can run the `personal-media-vault` binary.

In order to see the options, use:

```
personal-media-vault --help
```

In order to run the daemon, use:

```
personal-media-vault --daemon [OPTIONS]
```
