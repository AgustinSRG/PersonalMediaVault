use std::path::Path;

use pidlock::Pidlock;
use slint::Weak;

use crate::{
    log_debug,
    utils::{CancellableTaskController, LOCK_FILE_NAME},
    worker::tasks::backup::{
        check::check_backup_files, copy::copy_backup_files, finder::find_files_to_backup,
        progress::BackupProgressStatus,
    },
    MainWindow, VaultBackupCurrentTask, VaultBackupErrorType,
};

pub fn make_backup(
    window_handle: Weak<MainWindow>,
    task: &CancellableTaskController,
    vault_path: &str,
    backup_path: &str,
) -> Result<(), ()> {
    let mut progress = BackupProgressStatus::new(window_handle);

    if task.is_cancelled() {
        progress.set_idle();
        return Err(());
    }

    // Find files

    progress.start_task(VaultBackupCurrentTask::FindingFiles, true);

    let files = find_files_to_backup(task, &mut progress, vault_path)?;

    progress.end_task();

    if task.is_cancelled() {
        progress.set_idle();
        return Err(());
    }

    // Lock backup path

    let mut p = Path::new(backup_path).to_path_buf();
    p.push(LOCK_FILE_NAME);

    let mut lock = match Pidlock::new_validated(p) {
        Ok(l) => l,
        Err(e) => {
            log_debug!("Error: {e}");

            progress.set_error(VaultBackupErrorType::Locked, e.to_string());

            return Err(());
        }
    };

    match lock.acquire() {
        Ok(()) => {}
        Err(pidlock::PidlockError::LockExists) => {
            log_debug!("[LOCK] Another instance is already running");

            progress.set_error(VaultBackupErrorType::Locked, "".to_string());

            return Err(());
        }
        Err(e) => {
            log_debug!("[LOCK] Failed to acquire lock: {}", e);

            progress.set_error(VaultBackupErrorType::Locked, e.to_string());

            return Err(());
        }
    }

    // Check files

    progress.start_task(VaultBackupCurrentTask::CheckingFiles, false);

    let checked_files = check_backup_files(task, &mut progress, vault_path, backup_path, files)?;

    progress.files_total = checked_files.len() as u64;
    progress.bytes_total = progress.bytes_done;

    progress.end_task();

    if task.is_cancelled() {
        progress.set_idle();
        return Err(());
    }

    // Copy files

    progress.start_task(VaultBackupCurrentTask::CopyingFiles, false);

    copy_backup_files(task, &mut progress, vault_path, backup_path, checked_files)?;

    progress.end_task();

    if task.is_cancelled() {
        progress.set_idle();
        return Err(());
    }

    progress.set_success();

    Ok(())
}
