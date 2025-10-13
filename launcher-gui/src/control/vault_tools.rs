// Callbacks for vault status view

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks_vault_tools(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_run_tool({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move |selected_tool| {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);

            let _ = worker_sender_c.send(LauncherWorkerMessage::RunTool {
                tool: selected_tool,
            });
        }
    });

    ui.on_cancel_tool({
        let ui_handle = ui.as_weak();
        let worker_sender_c = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);

            let _ = worker_sender_c.send(LauncherWorkerMessage::CancelTool);
        }
    });
}
