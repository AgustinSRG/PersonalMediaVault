use std::{fs, io::ErrorKind, path::Path, time::SystemTime};

use crate::{
    log_debug,
    utils::CancellableTaskController,
    worker::tasks::backup::{
        entry::{BackupEntry, CheckedBackupEntry},
        progress::BackupProgressStatus,
    },
    VaultBackupErrorType,
};

pub fn check_backup_files(
    task: &CancellableTaskController,
    progress: &mut BackupProgressStatus,
    vault_path: &str,
    backup_path: &str,
    files: Vec<BackupEntry>,
) -> Result<Vec<CheckedBackupEntry>, ()> {
    let mut result: Vec<CheckedBackupEntry> = Vec::new();

    let vault_path_buf = Path::new(vault_path).to_path_buf();
    let backup_path_buf = Path::new(backup_path).to_path_buf();

    for file in files {
        if progress.should_update() {
            if task.is_cancelled() {
                progress.set_idle();
                return Err(());
            } else {
                progress.update();
            }
        }

        // Check the metadata of the file

        let original_file_path = vault_path_buf.join(&file.path);

        let stat_original = match fs::metadata(&original_file_path) {
            Ok(m) => m,
            Err(err) => {
                match err.kind() {
                    ErrorKind::NotFound => {}
                    _ => {
                        log_debug!(
                            "Error (file: {}): {}",
                            original_file_path.to_string_lossy().to_string(),
                            err
                        );
                    }
                }

                progress.files_done += 1;
                continue;
            }
        };

        let file_size = stat_original.len();
        let modified_original = stat_original.modified().unwrap_or(SystemTime::now());

        // Check the backup dir

        let backup_file_path = backup_path_buf.join(&file.path);

        let should_update = match fs::metadata(&backup_file_path) {
            Ok(m) => match m.modified() {
                Ok(modified_backup) => modified_backup < modified_original,
                Err(err) => {
                    log_debug!(
                        "Error (file: {}): {}",
                        backup_file_path.to_string_lossy().to_string(),
                        err.to_string()
                    );
                    true
                }
            },
            Err(err) => match err.kind() {
                ErrorKind::NotFound => true,
                _ => {
                    log_debug!(
                        "Error (file: {}): {}",
                        backup_file_path.to_string_lossy().to_string(),
                        err.to_string()
                    );
                    progress.set_error(
                        VaultBackupErrorType::Unknown,
                        format!(
                            "Error (file: {}): {}",
                            backup_file_path.to_string_lossy().to_string(),
                            err.to_string()
                        ),
                    );
                    return Err(());
                }
            },
        };

        if !should_update {
            progress.files_done += 1;
            continue;
        }

        progress.files_done += 1;
        progress.bytes_done += file_size;

        result.push(CheckedBackupEntry::new(file, file_size, modified_original));
    }

    Ok(result)
}
