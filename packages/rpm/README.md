# RPM package builder

This is a script to build a RPM package for Fedora Linux.

## Requirements

- [Golang](https://go.dev/) - Last stable version
- [NodeJS](https://nodejs.org/) - Last stable version
- [Rust](https://www.rust-lang.org/) - Last stable version

The RPM dev trools are also required.

```sh
sudo dnf install fedora-packager rpmdevtools
rpmdev-setuptree
```

## Building

Crate the `rpm` package file with the following command:

```sh
./build.sh
```
