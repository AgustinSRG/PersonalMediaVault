// Open vault

use crate::{
    models::LauncherConfig,
    utils::{
        file_exists, folder_exists, get_launcher_config_file, load_launcher_config_from_file,
        lock_vault, remove_existing_lock,
    },
    worker::{LauncherWorkerMessage, WorkerThreadStatus},
    LauncherStatus, MainWindow, OpenErrorType,
};
use slint::Weak;
use std::{fs, path::Path, sync::mpsc::Sender};

pub fn try_open_vault(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
) {
    // Check if path exists as a folder

    if !folder_exists(&status.vault_path) {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_launcher_status(LauncherStatus::CreateAsk);
        });
        return;
    }

    let lock = match lock_vault(&status.vault_path) {
        Ok(l) => l,
        Err(_) => {
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_launcher_status(LauncherStatus::LockAsk);
            });
            return;
        }
    };

    status.vault_lock = Some(lock);

    open_vault(status, sender, window_handle);
}

pub fn create_folder_and_open(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
) {
    if let Err(e) = fs::create_dir_all(&status.vault_path) {
        eprintln!("Error: {e}");
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_launcher_status(LauncherStatus::OpenError);
            win.set_open_error_type(OpenErrorType::CreateFolderError);
        });
        return;
    }

    try_open_vault(status, sender, window_handle);
}

pub fn force_open_vault(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
) {
    remove_existing_lock(&status.vault_path);

    try_open_vault(status, sender, window_handle);
}

pub fn open_vault(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
) {
    status.launcher_config_file = get_launcher_config_file(&status.vault_path);

    status.launcher_config = match load_launcher_config_from_file(&status.launcher_config_file) {
        Ok(c) => c,
        Err(_) => LauncherConfig::new(),
    };

    if status.launcher_config.port == 0 {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_hostname("localhost".into());
            win.set_hostname_invalid(false);
            win.set_port("8000".into());
            win.set_port_invalid(false);
            win.set_listen_local(true);
            win.set_launcher_status(LauncherStatus::InitialConfig);
        });
        return;
    }

    let mut credentials_file = Path::new(&status.vault_path).to_path_buf();
    credentials_file.push("credentials.json");

    if !file_exists(credentials_file) {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_launcher_status(LauncherStatus::CreateVaultAsk);
        });
        return;
    }

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_launcher_status(LauncherStatus::Open);
    });
    return;
}
