// Callbacks for the vault closed state

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks_vault_closed(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
     ui.on_open_default_vault({
        move || {
            println!("TODO");
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
