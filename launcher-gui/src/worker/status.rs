// Worker thread status

use pidlock::Pidlock;

use crate::models::{FFmpegConfig, LauncherConfig};

pub struct WorkerThreadStatus {
    pub daemon_binary: String,

    pub ffmpeg_config: FFmpegConfig,

    pub vault_path: String,
    pub vault_lock: Option<Pidlock>,

    pub launcher_config_file: String,
    pub launcher_config: LauncherConfig,
}

impl WorkerThreadStatus {
    pub fn new(daemon_binary: String, ffmpeg_config: FFmpegConfig) -> WorkerThreadStatus {
        WorkerThreadStatus {
            daemon_binary: daemon_binary,
            ffmpeg_config: ffmpeg_config,

            vault_path: "".to_string(),
            vault_lock: None,

            launcher_config_file: "".to_string(),
            launcher_config: LauncherConfig::default(),
        }
    }
}
