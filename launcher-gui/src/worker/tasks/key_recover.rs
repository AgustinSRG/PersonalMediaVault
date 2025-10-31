use std::{path::Path, sync::mpsc::Sender};

use slint::Weak;

use crate::{
    models::VaultCredentials,
    utils::{test_encryption_key_in_vault, EncryptionKeyTestResult},
    worker::{run_vault, stop_vault, LauncherWorkerMessage, WorkerThreadStatus},
    MainWindow, VaultBackupErrorType, VaultBackupStatus,
};

pub struct RecoverEncryptionKeyOptions {
    pub key: Vec<u8>,
    pub username: String,
    pub password: String,
}

pub fn recover_encryption_key(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    options: RecoverEncryptionKeyOptions,
) {
    // To do this we need to stop the vault first

    stop_vault(status, window_handle);

    // Test the encryption key

    let test_res = test_encryption_key_in_vault(&status.vault_path, &options.key);

    match test_res {
        EncryptionKeyTestResult::Valid => {}
        EncryptionKeyTestResult::Invalid => {
            run_vault(status, sender, window_handle, false);

            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_busy(false);
                win.set_backup_status(VaultBackupStatus::Idle);
                win.set_encryption_key_invalid(true);
            });
            return;
        }
        EncryptionKeyTestResult::NotEncryptedFiles => {
            run_vault(status, sender, window_handle, false);

            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_busy(false);
                win.set_backup_status(VaultBackupStatus::Error);
                win.set_backup_error_type(VaultBackupErrorType::NoEncryptedFiles);
            });
            return;
        }
    }

    // Key is valid, create new credentials file

    let mut credentials_path = Path::new(&status.vault_path).to_path_buf();
    credentials_path.push("credentials.json");

    let mut vault_credentials = match VaultCredentials::load_from_file(&credentials_path) {
        Ok(c) => c,
        Err(err) => {
            run_vault(status, sender, window_handle, false);

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

    vault_credentials.user = options.username.clone();

    if let Err(err) = vault_credentials.recover_key(&options.key, &options.password) {
        run_vault(status, sender, window_handle, false);

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

    if let Err(err) = vault_credentials.write_to_file(&credentials_path) {
        run_vault(status, sender, window_handle, false);

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

    // Restart the vault

    run_vault(status, sender, window_handle, false);

    // Set the success status

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_busy(false);
        win.set_backup_status(VaultBackupStatus::Success);
    });
}
