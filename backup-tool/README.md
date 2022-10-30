# Personal media vault (backup tool)

This is a backup tool for Personal Media Vault. Works similar to `rsync`, which can be used alternatively.

## Compilation

In order to install dependencies, type:

```
go get github.com/AgustinSRG/PersonalMediaVault/backup-tool
```

To compile the code type:

```
go build -o pmv-backup
```

The build command will create a binary in the current directory, called `pmv-backup`, or `pmv-backup.exe` if you are using Windows.

## Usage

The tool takes as the first argument the path of the vault, and the second argument the path of the backup folder.

```
pmv-backup </PATH/TO/VAULT> </PATH/TO/BACKUP>
```
