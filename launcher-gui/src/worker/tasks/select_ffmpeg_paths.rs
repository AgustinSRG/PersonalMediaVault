// Task to select a binary

use std::thread;

use slint::Weak;

use crate::MainWindow;

pub fn select_ffmpeg_binary(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new().add_filter("binary", &["", "exe"]);
        match dialog.pick_file() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_ffmpeg_path(path_cloned.into());
                        win.set_ffmpeg_path_invalid(false);
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

pub fn select_ffprobe_binary(window_handle: Weak<MainWindow>) {
    thread::spawn(move || {
        let dialog = rfd::FileDialog::new().add_filter("binary", &["", "exe"]);
        match dialog.pick_file() {
            Some(folder_path) => match folder_path.to_str() {
                Some(p) => {
                    let path_cloned = p.to_string();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = window_handle.unwrap();
                        win.set_busy(false);
                        win.set_ffprobe_path(path_cloned.into());
                        win.set_ffprobe_path_invalid(false);
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
