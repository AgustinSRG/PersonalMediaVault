use std::{process::Command, sync::mpsc::Sender};

use slint::Weak;

use crate::{
    worker::{tasks::open_vault, LauncherWorkerMessage, WorkerThreadStatus},
    LauncherStatus, MainWindow, OpenErrorType,
};

pub struct CreateVaultDetails {
    pub username: String,
    pub password: String,
}

pub fn create_vault(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    details: CreateVaultDetails,
) {
    let mut cmd = Command::new(&status.daemon_binary);

    cmd.args(["--init", "--skip-lock", "--vault-path", &status.vault_path]);

    cmd.env("PMV_INIT_SET_USER", details.username);
    cmd.env("PMV_INIT_SET_PASSWORD", details.password);

    if let Err(e) = cmd.output() {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_busy(false);
            win.set_launcher_status(LauncherStatus::OpenError);
            win.set_open_error_type(OpenErrorType::InitError);
            win.set_open_error_details(e.to_string().into());
        });

        return;
    }

    open_vault(status, sender, window_handle);
}
