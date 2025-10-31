// Callbacks for vault backup view

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{models::AES_KEY_SIZE, worker::LauncherWorkerMessage, MainWindow, VaultBackupStatus};

pub fn setup_callbacks_vault_backup(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_select_backup_path({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectBackupPath);
        }
    });

    ui.on_copy_encryption_key({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            let key = ui.get_encryption_key().to_string();
            let _ = sender.send(LauncherWorkerMessage::CopyToClipboard { contents: key });
        }
    });

    ui.on_run_backup({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move |option| {
            let ui = ui_handle.unwrap();

            match option {
                crate::VaultSelectedBackupOption::None => {}
                crate::VaultSelectedBackupOption::Backup => {}
                crate::VaultSelectedBackupOption::KeyExport => {
                    ui.set_username_invalid(false);
                    ui.set_password_invalid(false);

                    let mut has_error = false;

                    let username = ui.get_username().to_string();

                    if username.is_empty() {
                        has_error = true;
                        ui.set_username_invalid(true);
                    }

                    let password = ui.get_password().to_string();

                    if password.is_empty() {
                        has_error = true;
                        ui.set_password_invalid(true);
                    }

                    if has_error {
                        return;
                    }

                    ui.set_backup_status(VaultBackupStatus::Running);
                    ui.set_backup_progress_indeterminate(true);
                    ui.set_backup_progress_global("".to_string().into());
                    ui.set_backup_progress_file("".to_string().into());
                    ui.set_busy(true);
                    let _ =
                        sender.send(LauncherWorkerMessage::ExportVaultKey { username, password });
                }
                crate::VaultSelectedBackupOption::KeyRecover => {
                    ui.set_encryption_key_invalid(false);
                    ui.set_username_invalid(false);
                    ui.set_password_invalid(false);
                    ui.set_password_repeat_invalid(false);

                    let mut has_error = false;

                    let encryption_key =
                        hex::decode(ui.get_encryption_key().to_string()).unwrap_or(Vec::new());

                    if encryption_key.len() != AES_KEY_SIZE {
                        has_error = true;
                        ui.set_encryption_key_invalid(true);
                    }

                    let username = ui.get_username().to_string();

                    if username.is_empty() {
                        has_error = true;
                        ui.set_username_invalid(true);
                    }

                    let password = ui.get_password().to_string();
                    let password_repeat = ui.get_password_repeat().to_string();

                    if password.is_empty() {
                        has_error = true;
                        ui.set_password_invalid(true);
                    } else if password != password_repeat {
                        has_error = true;
                        ui.set_password_repeat_invalid(true);
                    }

                    if has_error {
                        return;
                    }

                    ui.set_backup_status(VaultBackupStatus::Running);
                    ui.set_backup_progress_indeterminate(true);
                    ui.set_backup_progress_global("".to_string().into());
                    ui.set_backup_progress_file("".to_string().into());
                    ui.set_busy(true);
                    let _ = sender.send(LauncherWorkerMessage::RecoverEncryptionKey {
                        key: encryption_key,
                        username,
                        password,
                    });
                }
            }
        }
    });
}
