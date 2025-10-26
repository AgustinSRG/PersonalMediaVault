// Control code for the UI

mod about;
use about::*;

mod vault_closed;
use vault_closed::*;

mod opening;
use opening::*;

mod vault_status;
use vault_status::*;

mod vault_config;
use vault_config::*;

mod vault_backup;
use vault_backup::*;

mod vault_tools;
use vault_tools::*;

use crate::{worker::LauncherWorkerMessage, MainWindow};
use std::sync::mpsc::Sender;

pub fn setup_callbacks(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    setup_callbacks_about(ui);
    setup_callbacks_vault_closed(ui, worker_sender.clone());
    setup_callbacks_vault_opening(ui, worker_sender.clone());
    setup_callbacks_vault_status(ui, worker_sender.clone());
    setup_callbacks_vault_config(ui, worker_sender.clone());
    setup_callbacks_vault_backup(ui, worker_sender.clone());
    setup_callbacks_vault_tools(ui, worker_sender);

    ui.on_close_launcher({
        move || {
            let _ = slint::quit_event_loop();
        }
    });
}
