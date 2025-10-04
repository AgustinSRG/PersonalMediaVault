// Worker thread status

use pidlock::Pidlock;

use crate::models::FFmpegConfig;

pub struct WorkerThreadStatus {
    pub ffmpeg_config: FFmpegConfig,
    pub vault_path: String,
    pub vault_lock: Option<Pidlock>,
}

impl WorkerThreadStatus {
    pub fn new(ffmpeg_config: FFmpegConfig) -> WorkerThreadStatus {
        WorkerThreadStatus{
            ffmpeg_config: ffmpeg_config,
            vault_path: "".to_string(),
            vault_lock: None,
        }
    }
}