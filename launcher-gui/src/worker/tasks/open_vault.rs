// Open vault

use crate::{
    utils::{folder_exists, lock_vault, remove_existing_lock},
    worker::{LauncherWorkerMessage, WorkerThreadStatus},
    FatalErrorType, LauncherStatus, MainWindow, OpenErrorType,
};
use slint::Weak;
use std::{fs, sync::mpsc::Sender};

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
    // TODO
    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_launcher_status(LauncherStatus::Open);
        win.set_busy(false);
    });
    return;
}
