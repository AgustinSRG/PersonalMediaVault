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

## Updating locales

This CLI program uses [go-i18n](https://github.com/nicksnyder/go-i18n) to support multiple languages.

You can install the tool to extract the messages from the code with the following command:

```
go install -v github.com/nicksnyder/go-i18n/v2/goi18n@latest
```

The messages files are named like `active.{LANG}.toml`

## Usage

In order to run the project, you can run the `pmv` binary.

In order to run it, use it with the vault path as its only argument:

```
pmv [PATH]
```
