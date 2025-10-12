// Task to select TLS cert and key

use std::thread;

use slint::Weak;

use crate::MainWindow;

pub fn select_tls_cert(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new().add_filter("pem", &["pem", "crt", "cer"]);
        match dialog.pick_file() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_tls_cert(path_cloned.into());
                        win.set_tls_cert_invalid(false);
                        win.set_dirty_tls(true);
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

pub fn select_tls_key(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new().add_filter("pem", &["pem", "key"]);
        match dialog.pick_file() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_tls_key(path_cloned.into());
                        win.set_tls_key_invalid(false);
                        win.set_dirty_tls(true);
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
