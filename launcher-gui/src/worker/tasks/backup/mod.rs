mod check;
mod copy;
mod entry;
mod finder;
mod make_backup;
mod progress;
mod thread;

use std::sync::mpsc::Sender;

use slint::Weak;

use crate::{
    utils::CancellableTaskController,
    worker::{tasks::backup::thread::run_backup_thread, LauncherWorkerMessage, WorkerThreadStatus},
    MainWindow,
};

pub fn start_backup(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    backup_path: String,
) {
    let task_id = status.backup_task_id;
    status.backup_task_id = status.backup_task_id.wrapping_add(1);

    let cancellable_task = CancellableTaskController::new();

    run_backup_thread(
        window_handle.clone(),
        sender.clone(),
        cancellable_task.clone(),
        task_id,
        status.vault_path.clone(),
        backup_path,
    );

    status.backup_cancellable_task = Some(cancellable_task);
}
