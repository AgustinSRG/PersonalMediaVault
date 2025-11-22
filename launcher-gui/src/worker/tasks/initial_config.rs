use std::sync::mpsc::Sender;

use slint::Weak;

use crate::{
    log_debug,
    utils::write_launcher_to_config_file,
    worker::{tasks::open_vault, LauncherWorkerMessage, WorkerThreadStatus},
    LauncherStatus, MainWindow, OpenErrorType,
};

pub struct InitialConfig {
    pub hostname: String,
    pub port: u16,
    pub local: bool,
}

pub fn set_initial_config(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    config: InitialConfig,
) {
    status.launcher_config.hostname = config.hostname;
    status.launcher_config.port = config.port;
    status.launcher_config.local = config.local;

    if let Err(e) =
        write_launcher_to_config_file(&status.launcher_config_file, &status.launcher_config)
    {
        log_debug!("Error: {e}");

        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_busy(false);
            win.set_launcher_status(LauncherStatus::OpenError);
            win.set_open_error_type(OpenErrorType::SaveConfigError);
            win.set_open_error_details(e.into());
        });

        let _ = sender.send(LauncherWorkerMessage::CloseVault);

        return;
    }

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_busy(false);
        win.set_launcher_status(LauncherStatus::Opening);
    });

    open_vault(status, sender, window_handle);
}
