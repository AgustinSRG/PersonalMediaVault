# Personal Media Vault

Web application to store and visualize media files (video, audio and pictures), meant to be run on a personal server or computer.

## Project motivation

This is a personal project developed with the goal of making a media management tool that has a simimilar web interface to Youtube, while keeping the media files encrypted in disk.

**Why encryption?:** When storing your personal pictures, videos or audio recordings, if you store them unencrypted in your computer, any malware that is able to infect your device will be able to easily get them from the file system. By encrypting them, it makes harder for those media files to be stolen. Also, by being encrypted, you can easily make backups of them in cloud storage services like Google Drive, without giving Google the ability to peek into your personal media files.

**Why a web application?:** By making it a web application, its usage becomes very flexible. You can to run the backend in a computer in your LAN or a remote server and you are able to use the app from all the devices connected to the Internet without any extra installation. Also, web interfaces are easier to develop to work in multiple devices and they offer standard features already implemented that are required for the app, like the video player.

## Features

 - Support for videos, audios and pictures.
 - Web interface: The project provides a Web interface, allowing the access to the vault from multiple devices and operating systems.
 - Encrypted and easy to back-up storage: The vault is encrypted using the user's password, and is stored in files in a data folder. It's possible to make backups of the valut just by copying the folder, ideal for tools like rsync. This project also provides its own backup tool for doing that.
 - Indexed tags: Each media asset can be tagged, so the user can search for it with a tag based search system.
 - Albums: Media files can be sorted using albums.
 - Media encoding: Media files are encoded into multiple formats in order to allow the playback from multiple browsers. For videos and pictures, they can also be resized into multiple resolutions for different kind of devices.
 - Automated thumbnail generation.
 - Subtitles support for video and audios.
 - Video timeline previews.

## Installation

Depending on the operating system you are using, there are multiple installation options.

### Windows

TODO

### Linux with DEB package support

TODO

## Project components

 - [Backend](./backend): Manages the vault, encodes and stores the media files and provides a HTTP API for clients to access it.
 - [Frontend](./frontend): Uses the HTTP API to access the vault, providing a web interface to the user.
 - [Backup tool](./backup-tool): Tool to make backups of media vaults (copies the new files and replaces the old ones, using the last modified date).
 - [Launcher](./launcher): CLI program to launch the web app. This component is made in order to make it easier to use in local, when you do not have the backend configured as a system service.
 - [Installation Packages](./packages): Collection of installation packages for multiple operating systems.

## License

This project is under the [MIT License](./LICENSE).
