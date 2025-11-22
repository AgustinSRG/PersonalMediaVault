// File utils

use std::{
    env, fs,
    path::{Path, PathBuf},
};

/// Checks if a file exists
pub fn file_exists<P>(file: P) -> bool
where
    P: AsRef<Path>,
{
    match fs::metadata(file) {
        Ok(m) => m.is_file(),
        Err(_) => false,
    }
}

/// Checks if a folder exists
pub fn folder_exists<P>(folder: P) -> bool
where
    P: AsRef<Path>,
{
    match fs::metadata(folder) {
        Ok(m) => m.is_dir(),
        Err(_) => false,
    }
}

/// Gets dirname of the current executable
pub fn get_dirname() -> PathBuf {
    let exe = match env::current_exe() {
        Ok(e) => e,
        Err(_) => {
            return Path::new("").to_path_buf();
        }
    };

    let dir = exe.parent().unwrap_or(Path::new(""));

    dir.to_path_buf()
}

/// Gets the name of a binary file
/// Appends ".exe" for Windows
pub fn get_binary_name(name: &str) -> String {
    if env::consts::OS == "windows" {
        name.to_string() + ".exe"
    } else {
        name.to_string()
    }
}
