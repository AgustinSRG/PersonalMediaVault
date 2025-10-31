// Utils to check an encryption key

use std::{
    fs::{self, DirEntry, File},
    io::Read,
    path::{Path, PathBuf},
};

use pmv_encryption_rs::decrypt;
use serde_json::value::RawValue;

use crate::log_debug;

pub enum EncryptionKeyTestResult {
    Valid,
    Invalid,
    NotEncryptedFiles,
}

const TEST_FILES_COUNT: usize = 3;

/// Tests encryption key with the vault files
pub fn test_encryption_key_in_vault(vault_path: &str, key: &[u8]) -> EncryptionKeyTestResult {
    let path = Path::new(vault_path);
    let mut test_files: Vec<PathBuf> = Vec::with_capacity(TEST_FILES_COUNT);

    find_pmv_files(&mut test_files, &path, TEST_FILES_COUNT);

    if test_files.is_empty() {
        return EncryptionKeyTestResult::NotEncryptedFiles;
    }

    for test_file in test_files {
        if !test_pmv_file(test_file, key) {
            return EncryptionKeyTestResult::Invalid;
        }
    }

    return EncryptionKeyTestResult::Valid;
}

fn find_pmv_files(result: &mut Vec<PathBuf>, path: &Path, limit: usize) {
    let files: Vec<DirEntry> = match fs::read_dir(path) {
        Ok(f) => f.flatten().collect(),
        Err(err) => {
            let path_str = path.to_string_lossy().to_string();
            log_debug!("Error (read_dir({path_str})): {err}");
            return;
        }
    };

    for file in files {
        if result.len() >= limit {
            return;
        }

        let file_path = file.path();
        let file_path_str = file_path.to_string_lossy().to_string();

        match file.file_type() {
            Ok(file_type) => {
                if file_type.is_file() && file_path_str.ends_with(".pmv") {
                    log_debug!("Test file: {file_path_str}");
                    result.push(file_path);
                } else if file_type.is_dir() {
                    find_pmv_files(result, &file_path, limit);
                }
            }
            Err(err) => {
                log_debug!("Error (stat({file_path_str})): {err}");
            }
        }
    }
}

fn test_pmv_file<T>(path: T, key: &[u8]) -> bool
where
    T: AsRef<Path>,
{
    // Read file

    let mut f = match File::open(path) {
        Ok(fh) => fh,
        Err(err) => {
            log_debug!("Error: {err}");
            return false;
        }
    };

    let mut bytes: Vec<u8> = Vec::new();

    if let Err(e) = f.read_to_end(&mut bytes) {
        log_debug!("Error: {e}");
        return false;
    }

    // Decrypt

    let decrypted_bytes = match decrypt(&bytes, key) {
        Ok(d) => d,
        Err(err) => {
            log_debug!("Error: {err}");
            return false;
        }
    };

    // Test JSON

    if let Err(e) = serde_json::from_slice::<&RawValue>(&decrypted_bytes) {
        log_debug!("Error: {e}");
        return false;
    };

    // All OK

    true
}
