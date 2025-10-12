use crate::{
    utils::{get_log_file, open_url_async},
    worker::{LauncherWorkerMessage, WorkerThreadStatus},
    MainWindow, VaultDaemonErrorType, VaultDaemonStatus,
};
use duct::{cmd, Handle};
use slint::Weak;
use std::{
    fs::File,
    process,
    sync::{
        mpsc::{channel, Sender},
        Arc,
    },
    thread::{self, sleep},
    time::{Duration, SystemTime, UNIX_EPOCH},
};

pub fn stop_vault(status: &mut WorkerThreadStatus, window_handle: &Weak<MainWindow>) {
    {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_daemon_status(VaultDaemonStatus::Stopping);
        });
    }

    if let Some(p) = &status.daemon_process {
        let _ = p.kill();
    }

    if let Some(r) = &status.daemon_process_wait_receiver {
        let _ = r.recv();
        status.daemon_process_wait_receiver = None;
    }
}

fn get_random_launcher_tag() -> String {
    let start = SystemTime::now();
    let since_the_epoch = start
        .duration_since(UNIX_EPOCH)
        .expect("time should go forward");

    let timestamp = since_the_epoch.as_millis();

    let pid = process::id();

    return timestamp.to_string() + "-" + &pid.to_string();
}

pub fn run_vault(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
) {
    stop_vault(status, window_handle); // In case the vault is not stopped, stop it

    {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_daemon_status(VaultDaemonStatus::Starting);
        });
    }

    let log_file_path = match get_log_file() {
        Ok(p) => p,
        Err(e) => {
            let err_str = e.to_string();
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_daemon_status(VaultDaemonStatus::Error);
                win.set_daemon_error_type(VaultDaemonErrorType::Unknown);
                win.set_daemon_error_details(err_str.into());
            });
            return;
        }
    };

    let log_file = match File::create(&log_file_path) {
        Ok(f) => f,
        Err(e) => {
            let err_str = e.to_string();
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_daemon_status(VaultDaemonStatus::Error);
                win.set_daemon_error_type(VaultDaemonErrorType::Unknown);
                win.set_daemon_error_details(err_str.into());
            });
            return;
        }
    };

    let launcher_tag = get_random_launcher_tag();

    let bind_addr = if status.launcher_config.local {
        "127.0.0.1".to_string()
    } else {
        "".to_string()
    };

    let mut args: Vec<String> = vec![
        "--daemon".to_string(),
        "--skip-lock".to_string(),
        "--clean".to_string(),
        "--vault-path".to_string(),
        status.vault_path.clone(),
        "--port".to_string(),
        status.launcher_config.port.to_string(),
        "--bind".to_string(),
        bind_addr,
        "--launch-tag".to_string(),
        launcher_tag.clone(),
        "--cache-size".to_string(),
        status.launcher_config.cache_size.as_i32().to_string(),
    ];

    if status.launcher_config.log_requests {
        args.push("--log-requests".to_string());
    }

    if status.launcher_config.debug {
        args.push("--debug".to_string());
    }

    let cmd = cmd(status.daemon_binary.clone(), args)
        .unchecked()
        .env("FFMPEG_PATH", status.ffmpeg_config.ffmpeg_path.clone())
        .env("FFPROBE_PATH", status.ffmpeg_config.ffprobe_path.clone())
        .env(
            "FFMPEG_VIDEO_ENCODER",
            status.ffmpeg_config.video_codec.clone(),
        )
        .env("FRONTEND_PATH", status.frontend_path.clone())
        .env(
            "SSL_CERT",
            if status.launcher_config.has_ssl() {
                status.launcher_config.ssl_cert.clone()
            } else {
                "".to_string()
            },
        )
        .env(
            "SSL_KEY",
            if status.launcher_config.has_ssl() {
                status.launcher_config.ssl_key.clone()
            } else {
                "".to_string()
            },
        )
        .stderr_to_stdout()
        .stdout_file(log_file)
        .stdin_null();

    let handle = match cmd.start() {
        Ok(p) => p,
        Err(e) => {
            let err_str = e.to_string();
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_daemon_status(VaultDaemonStatus::Error);
                win.set_daemon_error_type(VaultDaemonErrorType::Unknown);
                win.set_daemon_error_details(err_str.into());
            });
            return;
        }
    };

    let daemon_process = Arc::new(handle);

    let (daemon_process_wait_sender, daemon_process_wait_receiver) = channel::<bool>();

    status.daemon_id = status.daemon_id.wrapping_add(1);
    let daemon_id = status.daemon_id;

    let health_check_url = status.launcher_config.get_health_check_url(&launcher_tag);

    wait_for_daemon_process(
        sender.clone(),
        daemon_process.clone(),
        daemon_process_wait_sender,
        daemon_id,
        health_check_url,
    );

    status.daemon_process = Some(daemon_process);
    status.daemon_process_wait_receiver = Some(daemon_process_wait_receiver);
    status.log_file = Some(log_file_path);
}

pub fn wait_for_daemon_process(
    sender: Sender<LauncherWorkerMessage>,
    process: Arc<Handle>,
    daemon_process_wait_sender: Sender<bool>,
    daemon_id: u64,
    health_check_url: String,
) {
    thread::spawn(move || {
        let mut started = false;

        while !started {
            // Check of the process is dead

            match process.try_wait() {
                Ok(r) => {
                    if let Some(o) = r {
                        let status_code = o.status.code().unwrap_or(0);
                        match status_code {
                            4 => {
                                let _ = sender.send(LauncherWorkerMessage::VaultStartError {
                                    daemon_id,
                                    error_type: VaultDaemonErrorType::Lock,
                                    error_details: "".to_string(),
                                });
                            }
                            5 => {
                                let _ = sender.send(LauncherWorkerMessage::VaultStartError {
                                    daemon_id,
                                    error_type: VaultDaemonErrorType::PortInUse,
                                    error_details: "".to_string(),
                                });
                            }
                            6 => {
                                let _ = sender.send(LauncherWorkerMessage::VaultStartError {
                                    daemon_id,
                                    error_type: VaultDaemonErrorType::InvalidSsl,
                                    error_details: "".to_string(),
                                });
                            }
                            _ => {
                                let _ = sender.send(LauncherWorkerMessage::VaultStartError {
                                    daemon_id,
                                    error_type: VaultDaemonErrorType::Unknown,
                                    error_details: format!(
                                        "Daemon exit status code: {status_code}"
                                    ),
                                });
                            }
                        };
                        let _ = daemon_process_wait_sender.send(true);
                        return;
                    }
                }
                Err(e) => {
                    let _ = sender.send(LauncherWorkerMessage::VaultStartError {
                        daemon_id,
                        error_type: VaultDaemonErrorType::Unknown,
                        error_details: e.to_string(),
                    });
                    let _ = daemon_process_wait_sender.send(true);
                    return;
                }
            }

            if let Ok(res) = reqwest::blocking::get(&health_check_url) {
                if res.status() == 200 {
                    started = true;
                }
            };

            if !started {
                // Wait a few milliseconds to try again
                sleep(Duration::from_millis(100));
            }
        }

        let _ = sender.send(LauncherWorkerMessage::VaultStarted { daemon_id });

        let _ = sender.send(LauncherWorkerMessage::OpenBrowser);

        let _ = process.wait();

        let _ = daemon_process_wait_sender.send(true);

        let _ = sender.send(LauncherWorkerMessage::VaultStopped { daemon_id });
    });
}

pub fn on_vault_start_error(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    daemon_id: u64,
    error_type: VaultDaemonErrorType,
    error_details: String,
) {
    if status.daemon_id != daemon_id {
        return;
    }

    stop_vault(status, window_handle);

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_daemon_status(VaultDaemonStatus::Error);
        win.set_daemon_error_type(error_type);
        win.set_daemon_error_details(error_details.into());
    });
}

pub fn on_vault_started(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    daemon_id: u64,
) {
    if status.daemon_id != daemon_id {
        return;
    }

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_daemon_status(VaultDaemonStatus::Running);
    });
}

pub fn on_vault_stopped(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    daemon_id: u64,
) {
    if status.daemon_id != daemon_id {
        return;
    }

    stop_vault(status, window_handle);

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_daemon_status(VaultDaemonStatus::Stopped);
    });
}

pub fn open_vault_in_browser(status: &WorkerThreadStatus) {
    let browser_url = status.launcher_config.get_browser_url();
    open_url_async(&browser_url);
}

pub fn open_vault_log_file(status: &WorkerThreadStatus) {
    if let Some(f) = &status.log_file {
        open_url_async(f);
    }
}
