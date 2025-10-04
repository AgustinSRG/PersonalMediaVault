// Callbacks for the vault opening state

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, LauncherStatus, MainWindow};

pub fn setup_callbacks_vault_opening(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_create_folder_and_open({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_launcher_status(LauncherStatus::Opening);
            let _ = worker_sender_c.send(LauncherWorkerMessage::CreateFolderAndOpen);
        }
    });

    ui.on_force_open({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_launcher_status(LauncherStatus::Opening);
            let _ = worker_sender_c.send(LauncherWorkerMessage::ForceOpenVault);
        }
    });
}
