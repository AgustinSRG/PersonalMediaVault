// Background worker thread

use std::{
    sync::mpsc::{Receiver, Sender},
    thread::{spawn, JoinHandle},
};

mod message;
pub use message::*;

mod tasks;
use slint::{ComponentHandle, LogicalSize, Weak};
pub use tasks::*;

use crate::{
    models::{FFmpegBadInstallationError, FFmpegConfig},
    utils::load_ffmpeg_config,
    FatalErrorType, LauncherStatus, MainWindow,
};

/// Run the launcher main worker thread
pub fn run_worker_thread(
    sender: Sender<LauncherWorkerMessage>,
    receiver: Receiver<LauncherWorkerMessage>,
    window_handle: Weak<MainWindow>,
) -> JoinHandle<()> {
    // Spawn thread
    spawn(move || {
        let ffmpeg_config = match load_ffmpeg_config() {
            Ok(c) => c,
            Err(e) => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_launcher_status(LauncherStatus::FatalError);
                    match e {
                        FFmpegBadInstallationError::FFmpegMissing => {
                            win.set_fatal_error_type(FatalErrorType::FfmpegMissing);
                        }
                        FFmpegBadInstallationError::FFprobeMissing => {
                            win.set_fatal_error_type(FatalErrorType::FfprobeMissing);
                        }
                    }
                });

                FFmpegConfig::default_config()
            }
        };

        loop {
            match receiver.recv() {
                Ok(msg) => match msg {
                    LauncherWorkerMessage::OpenVault { path } => {
                        println!("Open vault: {path}");
                    }
                    LauncherWorkerMessage::SelectVaultFolder => {
                        select_vault_folder(sender.clone(), window_handle.clone());
                    }
                    LauncherWorkerMessage::Finish => {
                        return;
                    }
                },
                Err(err) => {
                    eprintln!("Error: {}", err);
                    return;
                }
            }
        }
    })
}
