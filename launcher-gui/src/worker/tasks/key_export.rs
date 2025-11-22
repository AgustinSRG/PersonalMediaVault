use std::path::Path;

use slint::Weak;

use crate::{
    models::VaultCredentials, worker::WorkerThreadStatus, MainWindow, VaultBackupErrorType,
    VaultBackupStatus,
};

pub fn export_vault_key(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    username: String,
    password: String,
) {
    let mut path = Path::new(&status.vault_path).to_path_buf();
    path.push("credentials.json");

    // Load vault credentials

    let vault_credentials = match VaultCredentials::load_from_file(path) {
        Ok(c) => c,
        Err(err) => {
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_busy(false);
                win.set_backup_status(VaultBackupStatus::Error);
                win.set_backup_error_type(VaultBackupErrorType::Unknown);
                win.set_backup_error(err.into());
            });
            return;
        }
    };

    if vault_credentials.user != username {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_busy(false);
            win.set_backup_status(VaultBackupStatus::Idle);
            win.set_username_invalid(true);
        });
        return;
    }

    let key = match vault_credentials.extract_encryption_key(&password) {
        Ok(k) => k,
        Err(err) => match err {
            crate::models::KeyExportError::InvalidPassword => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_busy(false);
                    win.set_backup_status(VaultBackupStatus::Idle);
                    win.set_password_invalid(true);
                });
                return;
            }
            crate::models::KeyExportError::CredentialsError(details) => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_busy(false);
                    win.set_backup_status(VaultBackupStatus::Error);
                    win.set_backup_error_type(VaultBackupErrorType::Unknown);
                    win.set_backup_error(details.into());
                });
                return;
            }
        },
    };

    let hey_hex = hex::encode(&key).to_uppercase();
    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_busy(false);
        win.set_backup_status(VaultBackupStatus::Success);
        win.set_encryption_key(hey_hex.into());
    });
}
