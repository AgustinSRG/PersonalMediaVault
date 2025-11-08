# Windows installer

Windows MSI installer for PersonalMediaVault

## Requirements

The following software is required to build the installer:

- [WiX ToolSet](https://wixtoolset.org/documentation/manual/v3/overview/alltools.html)
- [Golang](https://go.dev/) - Last stable version
- [NodeJS](https://nodejs.org/) - Last stable version
- [FFmpeg](https://ffmpeg.org/)
- [Rust](https://www.rust-lang.org/)

Make sure you have installed in `C:\ffmpeg\bin` the binaries `ffmpeg.exe` and `ffprobe.exe`.

## Building

To generate the MSI installer, run the following script:

```
build.bat
```
