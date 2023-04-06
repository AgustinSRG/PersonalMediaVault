# Personal Media Vault

Web application to store media files (video, audio and pictures) in an encrypted storage, and visualize them using a web browser.

## Project motivation

This is a personal project developed with the goal of making a media management tool that has a similar web interface to YouTube, while keeping the media files encrypted in disk.

**Why encryption?:** When storing your personal pictures, videos or audio recordings, if you store them unencrypted in your computer, any malware that is able to infect your device will be able to easily get them from the file system. By encrypting them, it makes harder for those media files to be stolen. Also, by being encrypted, you can easily make backups of them in cloud storage services like Google Drive, without giving Google the ability to peek into your personal media files.

**Why a web application?:** By making it a web application, its usage becomes very flexible. You can to run the backend in a computer in your LAN or a remote server and you are able to use the app from all the devices connected to the Internet without any extra installation. Also, web interfaces are easier to develop to work in multiple devices and they offer standard features already implemented that are required for the app, like the video player.

## Features

 - Support for videos, audios and pictures.
 - Web interface: The project provides a Web interface, allowing the access to the vault from multiple devices and operating systems.
 - Encrypted and easy to back-up storage: The vault is encrypted using the user's password, and is stored in files in a data folder. It's possible to make backups of the vault just by copying the folder, ideal for tools like rsync. This project also provides its own backup tool for doing that.
 - Indexed tags: Each media asset can be tagged, so the user can search for it with a tag based search system.
 - Albums: Media files can be sorted using albums.
 - Media encoding: Media files are encoded into multiple formats in order to allow the playback from multiple browsers. For videos and pictures, they can also be resized into multiple resolutions for different kind of devices.
 - Automated thumbnail generation.
 - Subtitles support for videos and audios.
 - Video timeline previews.
 - Video timeline slices.
 - Image annotations.

## Installation

Depending on the operating system you are using, there are multiple installation options.

### Linux with DEB package support

For apt-compatible Linux distributions, we provide a DEB package:

| Version | Arch | Hash | Hash alg. | Download |
|---|---|---|---|---|
| 1.4.1 | amd64 | `482d45619a2f8765e03b2bbefc4f2761296f25b6e773e894adbfe37831836dd3` | SHA256 | [Google Drive](https://drive.google.com/file/d/1P5LEZe1U9rX324NbLnns7tRZE4m5sv94/view?usp=sharing) / [Mega](https://mega.nz/file/BCllAKKQ#QxSCx1nfT51U9vOvbXDVxaFOJtJoRfVttiZDUgeItAk) |

Download it, and check its integrity using:

```sh
sh256sum ./personalmediavault_1.4-1.deb
```

Then, install it using `apt`:

```sh
sudo apt install ./personalmediavault_1.4-1.deb
```

After the installation, the binary files will be available in `/usr/bin`, and the frontend files will be available in `/usr/lib/pmv/www`.

To run a vault, type:

```sh
pmv /path/to/vault
```

You can change the language the launcher uses setting the `PMV_LANGUAGE` environment variable to `en` or `es`.

### Windows

For Windows, we provide a MSI installer:

| Version | Arch | Language | Hash | Download |
|---|---|---|---|---|
| 1.4.1 | x64 | English | SHA256: `CEBD1248E8BFE6B1FAA17AE72D4333BE9A5A04F7AF92FED7EB6EC6598FE484B8` | [Google Drive](https://drive.google.com/file/d/1PvEq_f3vNbrlTHZ4GkN2mnwS2LrCM4_P/view?usp=sharing) / [Mega](https://mega.nz/file/8D9BXC6I#72ggc-vJY9omr0jpBYlBeMsbwq6MTZaw6BN4CjVSs8g) |
| 1.4.1 | x64 | Spanish | SHA256: `5E64B8269AF05380BFBE5ADBEA1AB61C733BAC184A49ED1EA329238CEFBF9364` | [Google Drive](https://drive.google.com/file/d/1volAqFbWA9KyFlrJWviKnEi6spidgE9S/view?usp=sharing) / [Mega](https://mega.nz/file/IX0DyAwI#hTFdBhgD9cKV2kLYhhr97u6Lvkc2GYAsLFV108S5724) |

In order to install Personal Media Vault in Windows, run the installer  and everything will be set up for the application to work.

When downloading the MSI installer, check the file hash with PowerShell to make sure the file was not modified:

```ps1
Get-FileHash -Path "PersonalMediaVault-1.4.1-x64.msi" -Algorithm SHA256
Get-FileHash -Path "PersonalMediaVault-1.4.1-x64-es.msi" -Algorithm SHA256
```

After it's installed, create an empty folder to store your media vault and right click it. You should see a new option "Open with PersonalMediaVault". Click it to run the vault launcher.

## Docker

You can run this project with docker, using the [official image](https://hub.docker.com/r/asanrom/pmv) uploaded to Docker Hub.

In order to pull the image, type:

```
docker pull asanrom/pmv
```

To run a personal media vault instance, you can create a container, which is going to run the backend binary inside it.

Here is an example command to create a container:

```
docker run -p 80:80 -v /path/to/the/vault:/vault asanrom/pmv --daemon --clean --log-requests --vault-path /vault
```

You can replace `/path/to/the/vault` for the path where you have your vault stored.

For empty vaults, a default `admin`, with password `admin` will be created. You should change the password as soon as you first login into your vault to protect it with a strong password.

For more options, run:

```
docker run asanrom/pmv --help
```

## Project components

 - [Backend](./backend): Manages the vault, encodes and stores the media files and provides a HTTP API for clients to access it.
 - [Frontend](./frontend): Uses the HTTP API to access the vault, providing a web interface to the user.
 - [Backup tool](./backup-tool): Tool to make backups of media vaults (copies the new files and replaces the old ones, using the last modified date).
 - [Launcher](./launcher): CLI program to launch the web app. This component is made in order to make it easier to use in local, when you do not have the backend configured as a system service.
 - [Installation Packages](./packages): Collection of installation packages for multiple operating systems.

## License

This project is under the [MIT License](./LICENSE).
