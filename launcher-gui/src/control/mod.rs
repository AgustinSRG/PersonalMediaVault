// Control code for the UI

mod about;
use about::*;

mod vault_closed;
use vault_closed::*;

mod opening;
use opening::*;

use std::sync::mpsc::Sender;
use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    setup_callbacks_about(ui);
    setup_callbacks_vault_closed(ui, worker_sender.clone());
    setup_callbacks_vault_opening(ui, worker_sender);

    ui.on_close_launcher({
        move || {
            let _ = slint::quit_event_loop();
        }
    });
}
