use std::{
    fs::{self, File},
    io::{Read, Write},
    path::Path,
};

use crate::{
    log_debug,
    utils::CancellableTaskController,
    worker::tasks::backup::{entry::CheckedBackupEntry, progress::BackupProgressStatus},
    VaultBackupErrorType,
};

const READ_BUFFER_SIZE: usize = 8192;

pub fn copy_backup_files(
    task: &CancellableTaskController,
    progress: &mut BackupProgressStatus,
    vault_path: &str,
    backup_path: &str,
    files: Vec<CheckedBackupEntry>,
) -> Result<(), ()> {
    let vault_path_buf = Path::new(vault_path).to_path_buf();
    let backup_path_buf = Path::new(backup_path).to_path_buf();

    let mut temp_path = backup_path_buf.clone();
    temp_path.push("temp");
    let mut temp_file_index: u64 = 0;

    let _ = fs::create_dir_all(&temp_path);

    let mut read_buffer = [0_u8; READ_BUFFER_SIZE];

    for file in files {
        if progress.should_update() {
            if task.is_cancelled() {
                progress.set_idle();
                return Err(());
            } else {
                progress.update();
            }
        }

        let original_file_path = vault_path_buf.join(&file.path);
        let backup_file_path = backup_path_buf.join(&file.path);

        progress.start_file(file.path.to_string_lossy().to_string(), file.size);

        // Create temp file to copy

        temp_file_index = temp_file_index.wrapping_add(1);

        let mut temp_file_path = temp_path.clone();
        temp_file_path.push(format!("backup_tmp_{temp_file_index}"));

        let mut temp_file = match File::create(&temp_file_path) {
            Ok(f) => f,
            Err(err) => {
                log_debug!(
                    "Error (file: {}): {}",
                    temp_file_path.to_string_lossy().to_string(),
                    err.to_string()
                );
                progress.set_error(
                    VaultBackupErrorType::Unknown,
                    format!(
                        "Error {} creating file {}",
                        err,
                        temp_file_path.to_string_lossy()
                    ),
                );
                return Err(());
            }
        };

        // Read from the original and copy bytes

        let mut original_file = match File::open(&original_file_path) {
            Ok(f) => f,
            Err(err) => {
                log_debug!(
                    "Error (file: {}): {}",
                    original_file_path.to_string_lossy().to_string(),
                    err.to_string()
                );
                progress.set_error(
                    VaultBackupErrorType::Unknown,
                    format!(
                        "Error {} opening file {}",
                        err,
                        original_file_path.to_string_lossy()
                    ),
                );
                return Err(());
            }
        };

        loop {
            if progress.should_update() {
                if task.is_cancelled() {
                    progress.set_idle();
                    return Err(());
                } else {
                    progress.update();
                }
            }

            let size_read = match original_file.read(&mut read_buffer) {
                Ok(s) => s,
                Err(err) => {
                    log_debug!(
                        "Error (file: {}): {}",
                        original_file_path.to_string_lossy().to_string(),
                        err.to_string()
                    );
                    progress.set_error(
                        VaultBackupErrorType::Unknown,
                        format!(
                            "Error {} reading file {}",
                            err,
                            original_file_path.to_string_lossy()
                        ),
                    );
                    return Err(());
                }
            };

            if size_read == 0 {
                break;
            }

            if let Err(err) = temp_file.write_all(&read_buffer[..size_read]) {
                log_debug!(
                    "Error (file: {}): {}",
                    temp_file_path.to_string_lossy().to_string(),
                    err.to_string()
                );
                progress.set_error(
                    VaultBackupErrorType::Unknown,
                    format!(
                        "Error {} writing file {}",
                        err,
                        temp_file_path.to_string_lossy()
                    ),
                );
                return Err(());
            }

            progress.file_bytes_done += size_read as u64;
            progress.bytes_done += size_read as u64;
        }

        drop(original_file);

        // Set last modified data

        if let Err(err) = temp_file.set_modified(file.modified) {
            log_debug!(
                "Error (file: {}): {}",
                temp_file_path.to_string_lossy().to_string(),
                err.to_string()
            );
            progress.set_error(
                VaultBackupErrorType::Unknown,
                format!(
                    "Error {} setting modified date of file {}",
                    err,
                    temp_file_path.to_string_lossy()
                ),
            );
            return Err(());
        }

        // Correct file size

        if progress.file_bytes_done < file.size {
            progress.bytes_total -= file.size - progress.file_bytes_done;
        } else if progress.file_bytes_done > file.size {
            progress.bytes_total += progress.file_bytes_done - file.size;
        }

        progress.end_file();

        // Create folder (if applicable)

        if let Some(parent) = backup_file_path.parent() {
            let _ = fs::create_dir_all(parent);
        }

        // Move file

        if let Err(err) = fs::rename(&temp_file_path, &backup_file_path) {
            log_debug!(
                "Error (file: {}): {}",
                backup_file_path.to_string_lossy().to_string(),
                err.to_string()
            );
            progress.set_error(
                VaultBackupErrorType::Unknown,
                format!(
                    "Error {} moving file {}",
                    err,
                    backup_file_path.to_string_lossy()
                ),
            );
            return Err(());
        }

        // Increment file progress

        progress.files_done += 1;
    }

    Ok(())
}
