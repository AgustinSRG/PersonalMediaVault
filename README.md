# Personal Media Vault

**Project in development**

Web application to store and visualize media files (video, audio and pictures), meant to be run on a personal server or computer.

Features:

 - Encrypted and easy to back-up storage: The vault is encrypted using the user's password, and is stored in files in a data folder. It's possible to make backups of the valut just by copying the folder, ideal for tools like rsync.
 - Indexed tags: Each media asset can be tagged, so the user can search for it with a tag based search system.
 - Media encoding: Media files are encoded into multiple formats in order to alllow the playback from multiple browsers. For videos and pictures, they are also resized into multiple resolutions.
 - Web interface: The project provides a Web interface, allowing the access to the vault from multiple devices and operating systems.

Project components:

 - [Backend](./backend): Manages the vault, encodes the media and provides an HTTP API for clients to access it.
 - [Frontend](./frontend): Uses the HTTP API to access the vault, providing a web interface to the user.
