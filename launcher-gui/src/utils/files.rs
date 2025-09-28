// File utils

use std::fs;

/// Checks if a file exists
pub fn file_exists(file: &str) -> bool {
    match fs::metadata(file) {
        Ok(m) => m.is_file(),
        Err(_) => false,
    }
}
