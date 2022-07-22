# Windows installer

Windows MSI installer for PersonalMediaVault

## Requeriments

- [WiX Toolset](https://wixtoolset.org/documentation/manual/v3/overview/alltools.html)
- [FFMpeg](https://ffmpeg.org/)

## Building

First, make sure you have built all of PersonalMediaVault components (backend, frontend, backup-tool)

Also, make sure you have installed in `C:\ffmpeg\bin` the binaries `ffmpeg.exe` and `ffprobe.exe`

To copy the compiled files into a source folder, type:

```
powershell < configure.ps1
```

To generate the intaller, run the following script:

```
make-wix.bat
```
