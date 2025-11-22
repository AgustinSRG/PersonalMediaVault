// Callbacks for vault config view

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{
    models::FFmpegConfig,
    utils::{check_ffmpeg_codec, file_exists},
    worker::LauncherWorkerMessage,
    MainWindow,
};

pub fn setup_callbacks_vault_config(ui: &MainWindow, worker_sender: Sender<LauncherWorkerMessage>) {
    ui.on_select_ffmpeg_path({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectFFmpegBinary);
        }
    });

    ui.on_select_ffprobe_path({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectFFprobeBinary);
        }
    });

    ui.on_select_tls_cert({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectTlsCert);
        }
    });

    ui.on_select_tls_key({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::SelectTlsKey);
        }
    });

    ui.on_reset_config({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();
            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::ResetConfig);
        }
    });

    ui.on_update_port_host({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let local = ui.get_listen_local();
            let host = ui.get_hostname().to_string();

            if !hostname_validator::is_valid(&host) {
                ui.set_hostname_invalid(true);
                return;
            } else {
                ui.set_hostname_invalid(false);
            }

            let port = match ui.get_port().as_str().parse::<u16>() {
                Ok(p) => p,
                Err(_) => {
                    ui.set_port_invalid(true);
                    return;
                }
            };

            ui.set_port_invalid(false);

            ui.set_busy(true);
            let _ = sender.send(LauncherWorkerMessage::UpdateHostPortConfig { host, port, local });
        }
    });

    ui.on_update_tls({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let enabled = ui.get_tls_enabled();

            let cert = ui.get_tls_cert().to_string();
            let key = ui.get_tls_key().to_string();

            if enabled {
                if cert.is_empty() || !file_exists(&cert) {
                    ui.set_tls_cert_invalid(true);
                    return;
                } else {
                    ui.set_tls_cert_invalid(false);
                }

                if key.is_empty() || !file_exists(&key) {
                    ui.set_tls_key_invalid(true);
                    return;
                } else {
                    ui.set_tls_key_invalid(false);
                }
            } else {
                ui.set_tls_cert_invalid(false);
                ui.set_tls_key_invalid(false);
            }

            ui.set_busy(true);

            let _ = sender.send(LauncherWorkerMessage::UpdateTlsConfig { enabled, cert, key });
        }
    });

    ui.on_update_ffmpeg({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let ffmpeg_path = ui.get_ffmpeg_path().to_string();
            let ffprobe_path = ui.get_ffprobe_path().to_string();
            let video_codec = ui.get_video_codec().to_string();

            if ffmpeg_path.is_empty() || !file_exists(&ffmpeg_path) {
                ui.set_ffmpeg_path_invalid(true);
                return;
            } else {
                ui.set_ffmpeg_path_invalid(false);
            }

            if ffprobe_path.is_empty() || !file_exists(&ffprobe_path) {
                ui.set_ffprobe_path_invalid(true);
                return;
            } else {
                ui.set_ffprobe_path_invalid(false);
            }

            let conf = FFmpegConfig {
                ffmpeg_path: ffmpeg_path.clone(),
                ffprobe_path: ffprobe_path.clone(),
                video_codec: video_codec.clone(),
            };

            if video_codec.is_empty() || !check_ffmpeg_codec(&conf) {
                ui.set_video_codec_invalid(true);
                return;
            } else {
                ui.set_video_codec_invalid(false);
            }

            ui.set_busy(true);

            let _ = sender.send(LauncherWorkerMessage::UpdateFFmpegConfig {
                ffmpeg_path,
                ffprobe_path,
                video_codec,
            });
        }
    });

    ui.on_update_other({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let log_debug = ui.get_log_debug();
            let log_requests = ui.get_log_requests();

            let cache_size = match ui.get_cache_size().as_str().parse::<i32>() {
                Ok(p) => p,
                Err(_) => {
                    ui.set_cache_size_invalid(true);
                    return;
                }
            };

            ui.set_cache_size_invalid(false);

            ui.set_busy(true);

            let _ = sender.send(LauncherWorkerMessage::UpdateOtherConfig {
                cache_size,
                log_requests,
                log_debug,
            });
        }
    });
}
