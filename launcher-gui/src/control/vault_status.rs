// Callbacks for vault status view

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks_vault_status(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_start({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = worker_sender_c.send(LauncherWorkerMessage::StartVault);
        }
    });

    ui.on_restart({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = worker_sender_c.send(LauncherWorkerMessage::StartVault);
        }
    });

    ui.on_stop({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = worker_sender_c.send(LauncherWorkerMessage::StopVault);
        }
    });

    ui.on_open_browser({
        let worker_sender_c = worker_sender.clone();
        move || {
            let _ = worker_sender_c.send(LauncherWorkerMessage::OpenBrowser);
        }
    });

    ui.on_open_log({
        let worker_sender_c = worker_sender.clone();
        move || {
            let _ = worker_sender_c.send(LauncherWorkerMessage::OpenLogFile);
        }
    });
}
