use std::{
    fs::DirEntry,
    path::{Path, PathBuf},
};

use crate::{
    log_debug,
    utils::CancellableTaskController,
    worker::tasks::backup::{entry::BackupEntry, progress::BackupProgressStatus},
    VaultBackupErrorType,
};

fn find_files_recursive(
    task: &CancellableTaskController,
    entries: &mut Vec<BackupEntry>,
    progress: &mut BackupProgressStatus,
    vault_path: &PathBuf,
    path: PathBuf,
) -> Result<(), ()> {
    if progress.should_update() {
        if task.is_cancelled() {
            progress.set_idle();
            return Err(());
        } else {
            progress.update();
        }
    }

    let joined_path = vault_path.join(&path);

    match joined_path.as_path().read_dir() {
        Ok(f) => {
            let files_flatten: Vec<DirEntry> = f.flatten().collect();

            for file in files_flatten {
                let file_type = match file.file_type() {
                    Ok(t) => t,
                    Err(e) => {
                        log_debug!("Error: {e}");
                        continue;
                    }
                };

                let file_name = file.file_name().to_string_lossy().to_string();

                let mut new_path = path.clone();
                new_path.push(file_name);

                if file_type.is_file() {
                    entries.push(BackupEntry { path: new_path });
                    progress.files_done += 1;
                } else if file_type.is_dir() {
                    find_files_recursive(task, entries, progress, vault_path, new_path)?;
                }
            }
        }
        Err(err) => {
            progress.set_error(VaultBackupErrorType::Unknown, err.to_string());
            return Err(());
        }
    }

    Ok(())
}

fn find_files(
    task: &CancellableTaskController,
    entries: &mut Vec<BackupEntry>,
    progress: &mut BackupProgressStatus,
    vault_path: &str,
    relative_path: &[&str],
) -> Result<(), ()> {
    let vault_path_buf = Path::new(vault_path).to_path_buf();
    let relative_path_buf: PathBuf = relative_path.iter().collect();
    find_files_recursive(task, entries, progress, &vault_path_buf, relative_path_buf)
}

pub fn find_files_to_backup(
    task: &CancellableTaskController,
    progress: &mut BackupProgressStatus,
    vault_path: &str,
) -> Result<Vec<BackupEntry>, ()> {
    let mut entries: Vec<BackupEntry> = vec![
        BackupEntry::new(&["main.index"]),
        BackupEntry::new(&["credentials.json"]),
        BackupEntry::new(&["media_ids.json"]),
        BackupEntry::new(&["tasks.json"]),
        BackupEntry::new(&["albums.pmv"]),
        BackupEntry::new(&["tag_list.pmv"]),
        BackupEntry::new(&["user_config.pmv"]),
    ];

    progress.files_done = entries.len() as u64;

    find_files(task, &mut entries, progress, vault_path, &["tags"])?;
    find_files(task, &mut entries, progress, vault_path, &["media"])?;
    find_files(task, &mut entries, progress, vault_path, &["thumb_album"])?;

    Ok(entries)
}
