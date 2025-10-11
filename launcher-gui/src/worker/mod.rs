// Background worker thread

use std::{
    sync::mpsc::{Receiver, Sender},
    thread::{spawn, JoinHandle},
};

mod message;
pub use message::*;

mod tasks;
use slint::Weak;
pub use tasks::*;

mod status;
pub use status::*;

use crate::{
    log_debug,
    models::{FFmpegBadInstallationError, FFmpegConfig},
    utils::{find_pmv_daemon_binary, find_pmv_frontend, load_ffmpeg_config},
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
        let daemon_binary = match find_pmv_daemon_binary() {
            Ok(b) => b,
            Err(_) => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_launcher_status(LauncherStatus::FatalError);
                    win.set_fatal_error_type(FatalErrorType::DaemonMissing);
                });

                "pmvd".to_string()
            }
        };

        let frontend_path = match find_pmv_frontend() {
            Ok(b) => b,
            Err(_) => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_launcher_status(LauncherStatus::FatalError);
                    win.set_fatal_error_type(FatalErrorType::FrontendMissing);
                });

                "frontend".to_string()
            }
        };

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

        let mut status = WorkerThreadStatus::new(daemon_binary, frontend_path, ffmpeg_config);

        loop {
            match receiver.recv() {
                Ok(msg) => match msg {
                    LauncherWorkerMessage::OpenVault { path } => {
                        println!("Open vault: {path}");
                        status.vault_path = path;

                        try_open_vault(&mut status, &sender, &window_handle);
                    }
                    LauncherWorkerMessage::SelectVaultFolder => {
                        select_vault_folder(sender.clone(), window_handle.clone());
                    }
                    LauncherWorkerMessage::Finish => {
                        return;
                    }
                    LauncherWorkerMessage::CreateFolderAndOpen => {
                        create_folder_and_open(&mut status, &sender, &window_handle);
                    }
                    LauncherWorkerMessage::ForceOpenVault => {
                        force_open_vault(&mut status, &sender, &window_handle);
                    }
                    LauncherWorkerMessage::CloseVault => {
                        close_vault(&mut status);
                    }
                    LauncherWorkerMessage::SetInitialConfig {
                        hostname,
                        port,
                        local,
                    } => {
                        set_initial_config(
                            &mut status,
                            &sender,
                            &window_handle,
                            InitialConfig {
                                hostname: hostname,
                                port: port,
                                local: local,
                            },
                        );
                    }
                    LauncherWorkerMessage::CreateVault { username, password } => {
                        create_vault(
                            &mut status,
                            &sender,
                            &window_handle,
                            CreateVaultDetails {
                                username: username,
                                password: password,
                            },
                        );
                    }
                    LauncherWorkerMessage::StartVault => {
                        run_vault(&mut status, &sender, &window_handle);
                        let wh = window_handle.clone();
                        let _ = slint::invoke_from_event_loop(move || {
                            let win = wh.unwrap();
                            win.set_busy(false);
                        });
                    }
                    LauncherWorkerMessage::StopVault => {
                        stop_vault(&mut status, &window_handle);
                        let wh = window_handle.clone();
                        let _ = slint::invoke_from_event_loop(move || {
                            let win = wh.unwrap();
                            win.set_busy(false);
                        });
                    }
                    LauncherWorkerMessage::VaultStarted { daemon_id } => {
                        on_vault_started(&mut status, &window_handle, daemon_id);
                    }
                    LauncherWorkerMessage::VaultStartError {
                        daemon_id,
                        error_type,
                        error_details,
                    } => {
                        on_vault_start_error(
                            &mut status,
                            &window_handle,
                            daemon_id,
                            error_type,
                            error_details,
                        );
                    }
                    LauncherWorkerMessage::VaultStopped { daemon_id } => {
                        on_vault_stopped(&mut status, &window_handle, daemon_id);
                    }
                    LauncherWorkerMessage::OpenBrowser => {
                        open_vault_in_browser(&status);
                    }
                    LauncherWorkerMessage::OpenLogFile => {
                        open_vault_log_file(&status);
                    }
                },
                Err(err) => {
                    log_debug!("Error: {}", err);

                    return;
                }
            }
        }
    })
}
