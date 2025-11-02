use std::{path::{Path, PathBuf}, time::SystemTime};

/// Backup entry
/// A file to be backend up
pub struct BackupEntry {
    /// The file name
    pub path: PathBuf,
}

impl BackupEntry {
    pub fn new(file_path: &[&str]) -> BackupEntry {
        BackupEntry {
            path: file_path.iter().collect(),
        }
    }
}

/// Backup entry (checked)
pub struct CheckedBackupEntry {
    pub path: PathBuf,

    /// Size of the file
    pub size: u64,

    /// Modified time
    pub modified: SystemTime,
}

impl CheckedBackupEntry {
    pub fn new(entry: BackupEntry, size: u64, modified: SystemTime) -> CheckedBackupEntry {
        CheckedBackupEntry{
            path: entry.path,
            size,
            modified,
        }
    }
}
