# Personal Media Vault (CLI Launcher)

Launches a vault with an interactive command line interface to manage it.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/PersonalMediaVault/launcher
```

To compile the code type:

```
go build -o pmv
```

The build command will create a binary in the current directory, called `pmv`, or `pmv.exe` if you are using Windows.

## Usage

In order to run the project, you can run the `pmv` binary.

In order to run it, use it with the vault path as its only argument:

```
pmv [PATH]
```
