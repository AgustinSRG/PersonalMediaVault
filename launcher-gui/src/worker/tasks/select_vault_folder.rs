// Task to select a vault folder

use std::{sync::mpsc::Sender, thread};

use slint::Weak;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn select_vault_folder(sender: Sender<LauncherWorkerMessage>, window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new();
        match dialog.pick_folder() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                    });
                    let _ = sender.send(LauncherWorkerMessage::OpenVault { path: p.to_string() });
                }
                None => {
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                    });
                }
            },
            None => {
                let _ = slint::invoke_from_event_loop(move || {
                    let win = window_handle.unwrap();
                    win.set_busy(false);
                });
            }
        }
    });
}
