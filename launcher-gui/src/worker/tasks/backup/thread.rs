use std::{sync::mpsc::Sender, thread};

use slint::Weak;

use crate::{
    utils::CancellableTaskController,
    worker::{tasks::backup::make_backup::make_backup, LauncherWorkerMessage},
    MainWindow,
};

pub fn run_backup_thread(
    window_handle: Weak<MainWindow>,
    sender: Sender<LauncherWorkerMessage>,
    task: CancellableTaskController,
    task_id: u64,
    vault_path: String,
    backup_path: String,
) {
    thread::spawn(move || {
        let _ = make_backup(window_handle, &task, &vault_path, &backup_path);
        task.end();
        let _ = sender.send(LauncherWorkerMessage::BackupEnded { task_id });
    });
}
