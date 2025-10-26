// Callbacks for vault backup view

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks_vault_backup(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_select_backup_path({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectBackupPath);
        }
    });
}
