use std::sync::mpsc::Sender;

use slint::Weak;

use crate::{
    models::CacheSize,
    worker::{tasks::run_vault, LauncherWorkerMessage, WorkerThreadStatus},
    MainWindow,
};

pub fn reset_ui_config(status: &mut WorkerThreadStatus, window_handle: &Weak<MainWindow>) {
    {
        let wh = window_handle.clone();
        let launcher_config = status.launcher_config.clone();
        let ffmpeg_config = status.ffmpeg_config.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();

            win.set_tls_enabled(launcher_config.has_ssl());

            win.set_port(launcher_config.port.to_string().into());
            win.set_port_invalid(false);

            win.set_hostname(launcher_config.hostname.into());
            win.set_hostname_invalid(false);

            win.set_listen_local(launcher_config.local);

            win.set_tls_cert(launcher_config.ssl_cert.into());
            win.set_tls_cert_invalid(false);

            win.set_tls_key(launcher_config.ssl_key.into());
            win.set_tls_key_invalid(false);

            win.set_ffmpeg_path(ffmpeg_config.ffmpeg_path.into());
            win.set_ffmpeg_path_invalid(false);

            win.set_ffprobe_path(ffmpeg_config.ffprobe_path.into());
            win.set_ffmpeg_path_invalid(false);

            win.set_video_codec(ffmpeg_config.video_codec.into());
            win.set_video_codec_invalid(false);

            win.set_cache_size(launcher_config.cache_size.as_i32().to_string().into());
            win.set_cache_size_invalid(false);

            win.set_log_requests(launcher_config.log_requests);
            win.set_log_debug(launcher_config.debug);
        });
    }
}

pub struct HostPortConfigDetails {
    pub host: String,
    pub port: u16,
    pub local: bool,
}

pub fn update_config_host_port(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    details: HostPortConfigDetails,
) {
    status.launcher_config.hostname = details.host;
    status.launcher_config.port = details.port;
    status.launcher_config.local = details.local;

    status.save_launcher_config();

    run_vault(status, sender, window_handle);
}

pub struct TlsConfigDetails {
    pub enabled: bool,
    pub cert: String,
    pub key: String,
}

pub fn update_config_tls(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    details: TlsConfigDetails,
) {
    if details.enabled {
        status.launcher_config.ssl_cert = details.cert;
        status.launcher_config.ssl_key = details.key;
    } else {
        status.launcher_config.ssl_cert = "".to_string();
        status.launcher_config.ssl_key = "".to_string();
    }

    status.save_launcher_config();

    run_vault(status, sender, window_handle);
}

pub struct FFmpegConfigDetails {
    pub ffmpeg_path: String,
    pub ffprobe_path: String,
    pub video_codec: String,
}

pub fn update_config_ffmpeg(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    details: FFmpegConfigDetails,
) {
    status.ffmpeg_config.ffmpeg_path = details.ffmpeg_path;
    status.ffmpeg_config.ffprobe_path = details.ffprobe_path;
    status.ffmpeg_config.video_codec = details.video_codec;

    status.save_ffmpeg_config();

    run_vault(status, sender, window_handle);
}

pub struct OtherConfigDetails {
    pub cache_size: i32,
    pub log_requests: bool,
    pub log_debug: bool,
}

pub fn update_config_other(
    status: &mut WorkerThreadStatus,
    sender: &Sender<LauncherWorkerMessage>,
    window_handle: &Weak<MainWindow>,
    details: OtherConfigDetails,
) {
    status.launcher_config.cache_size = CacheSize::new(details.cache_size);
    status.launcher_config.log_requests = details.log_requests;
    status.launcher_config.debug = details.log_debug;

    status.save_launcher_config();

    run_vault(status, sender, window_handle);
}
