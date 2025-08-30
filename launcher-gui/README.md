# PersonalMediaVault - Launcher

This folder contains the source code of the launcher for PersonalMediaVault.

The launcher is a GUI app that facilitates running PersonalMediaVault in local desktop environments.

This app is made using [Slint](https://slint.dev/) for the interface, and [Rust](https://www.rust-lang.org/) as the programming language for the logic.

## Requirements

Make sure to install [Rust](https://www.rust-lang.org/) in your system.

Also, install the [slint-tr-extractor](https://crates.io/crates/slint-tr-extractor) tool, to extract translation messages from `.slint` files:

```sh
cargo install slint-tr-extractor
```

If you are in Windows, install the [GetText tools](https://gnuwin32.sourceforge.net/packages/gettext.htm), used to generate the translations files.

## Compiling

Use the compilation script to compile the app for production:

```sh
./build.sh
```

The result will be saved in `target/release`.

If you are compiling in Windows, use the `build.bat` script instead.

