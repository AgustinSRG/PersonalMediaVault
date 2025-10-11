// Callbacks for the vault closed state

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{
    utils::get_default_vault_path, worker::LauncherWorkerMessage, LauncherStatus, MainWindow,
};

pub fn setup_callbacks_vault_closed(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_open_default_vault({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            let vault_path = get_default_vault_path();
            ui.set_launcher_status(LauncherStatus::Opening);
            ui.set_vault_path(vault_path.clone().into());
            let _ = worker_sender_c.send(LauncherWorkerMessage::OpenVault { path: vault_path });
        }
    });

    ui.on_open_vault_folder({
        let ui_handle = ui.as_weak();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = worker_sender.send(LauncherWorkerMessage::SelectVaultFolder);
        }
    });
}
