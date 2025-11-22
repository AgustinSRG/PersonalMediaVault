// Task to select a path for the backup

use std::thread;

use slint::Weak;

use crate::MainWindow;

pub fn select_backup_path(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new();
        match dialog.pick_folder() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_backup_path(path_cloned.into());
                    });
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
