use slint::Weak;

use crate::{MainWindow, VaultBackupCurrentTask, utils::CancellableTaskController, worker::tasks::backup::{finder::find_files_to_backup, progress::BackupProgressStatus}};

pub fn make_backup(
    window_handle: Weak<MainWindow>,
    task: &CancellableTaskController,
    vault_path: &str,
    backup_path: &str,
) -> Result<(), ()> {
    let mut progress = BackupProgressStatus::new(window_handle);

    if task.is_cancelled()  {
        progress.set_idle();
        return Err(());
    }

    // Find files

    progress.start_task(VaultBackupCurrentTask::FindingFiles, true);

    let files = find_files_to_backup(task, &mut progress, vault_path)?;

    progress.end_task();

    if task.is_cancelled()  {
        progress.set_idle();
        return Err(());
    }

    // Check files

    progress.start_task(VaultBackupCurrentTask::CheckingFiles, false);



    progress.end_task();

    if task.is_cancelled()  {
        progress.set_idle();
        return Err(());
    }

    // Copy files

    progress.start_task(VaultBackupCurrentTask::CopyingFiles, false);



    progress.end_task();

    if task.is_cancelled()  {
        progress.set_idle();
        return Err(());
    }

    progress.set_success();

    Ok(())
}
