// Worker thread status

use std::{
    process::Child,
    sync::{
        mpsc::{Receiver, Sender},
        Arc,
    },
};

use duct::Handle;
use pidlock::Pidlock;

use crate::models::{FFmpegConfig, LauncherConfig};

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
}

impl WorkerThreadStatus {
    pub fn new(
        daemon_binary: String,
        frontend_path: String,
        ffmpeg_config: FFmpegConfig,
    ) -> WorkerThreadStatus {
        WorkerThreadStatus {
            daemon_binary: daemon_binary,
            frontend_path: frontend_path,
            ffmpeg_config: ffmpeg_config,

            vault_path: "".to_string(),
            vault_lock: None,

            launcher_config_file: "".to_string(),
            launcher_config: LauncherConfig::default(),

            daemon_id: 0,
            daemon_process: None,
            daemon_process_wait_receiver: None,

            log_file: None,
        }
    }
}
