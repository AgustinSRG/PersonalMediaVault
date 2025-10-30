// Worker thread status

use std::sync::{mpsc::Receiver, Arc};

use arboard::Clipboard;
use duct::Handle;
use pidlock::Pidlock;

use crate::{
    log_debug,
    models::{FFmpegConfig, LauncherConfig},
    utils::{get_clipboard, write_ffmpeg_to_config_file, write_launcher_to_config_file},
};

pub struct WorkerThreadStatus {
    pub daemon_binary: String,
    pub frontend_path: String,

    pub ffmpeg_config: FFmpegConfig,

    pub vault_path: String,
    pub vault_lock: Option<Pidlock>,

    pub launcher_config_file: String,
    pub launcher_config: LauncherConfig,

    pub daemon_id: u64,
    pub daemon_process: Option<Arc<Handle>>,
    pub daemon_process_wait_receiver: Option<Receiver<bool>>,

    pub log_file: Option<String>,

    pub tool_id: u64,
    pub tool_process: Option<Arc<Handle>>,
    pub tool_process_wait_receiver: Option<Receiver<bool>>,

    pub clipboard: Option<Clipboard>,
}

impl WorkerThreadStatus {
    pub fn new(
        daemon_binary: String,
        frontend_path: String,
        ffmpeg_config: FFmpegConfig,
    ) -> WorkerThreadStatus {
        WorkerThreadStatus {
            daemon_binary,
            frontend_path,
            ffmpeg_config,

            vault_path: "".to_string(),
            vault_lock: None,

            launcher_config_file: "".to_string(),
            launcher_config: LauncherConfig::default(),

            daemon_id: 0,
            daemon_process: None,
            daemon_process_wait_receiver: None,

            log_file: None,

            tool_id: 0,
            tool_process: None,
            tool_process_wait_receiver: None,

            clipboard: get_clipboard(),
        }
    }

    pub fn save_launcher_config(&self) {
        if let Err(e) =
            write_launcher_to_config_file(&self.launcher_config_file, &self.launcher_config)
        {
            log_debug!("Error: {e}");
        }
    }

    pub fn save_ffmpeg_config(&self) {
        if let Err(e) = write_ffmpeg_to_config_file(&self.ffmpeg_config) {
            log_debug!("Error: {e}");
        }
    }
}
