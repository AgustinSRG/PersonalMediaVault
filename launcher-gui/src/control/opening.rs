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

    ui.on_close_vault({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_launcher_status(LauncherStatus::Closed);
            let _ = worker_sender_c.send(LauncherWorkerMessage::CloseVault);
        }
    });

    ui.on_set_initial_config({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let local = ui.get_listen_local();
            let hostname = ui.get_hostname().to_string();

            if !hostname_validator::is_valid(&hostname) {
                ui.set_hostname_invalid(true);
                return;
            } else {
                ui.set_hostname_invalid(false);
            }

            let port = match ui.get_port().as_str().parse::<u16>() {
                Ok(p) => p,
                Err(_) => {
                    ui.set_port_invalid(true);
                    return;
                },
            };

            ui.set_port_invalid(false);

            ui.set_busy(true);
            let _ = worker_sender_c.send(LauncherWorkerMessage::SetInitialConfig { hostname: hostname, port: port, local: local });
        }
    });
}
