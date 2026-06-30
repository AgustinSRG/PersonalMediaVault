// Task to select a binary

use std::thread;

use slint::Weak;

use crate::MainWindow;

pub fn select_sse_model_path(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new();
        match dialog.pick_folder() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_model_path(path_cloned.into());
                        win.set_model_path_invalid(false);
                        win.set_dirty_sse(true);
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
