// FFmpeg config model

use serde::{Deserialize, Serialize};

pub const VIDEO_CODEC_DEFAULT: &str = "libx264";
pub const VIDEO_CODEC_ALTERNATIVE: &str = "libvpx-vp9";

/// Represents the FFmpeg configuration
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct FFmpegConfig {
    #[serde(default, rename = "ffmpeg_path")]
    pub ffmpeg_path: String,

    #[serde(default, rename = "ffprobe_path")]
    pub ffprobe_path: String,

    #[serde(default, rename = "video_codec")]
    pub video_codec: String,
}

impl FFmpegConfig {
    /// Gets a default configuration, in case of error loading the actual configuration
    pub fn default_config() -> FFmpegConfig {
        FFmpegConfig {
            ffmpeg_path: "/usr/bin/ffmpeg".to_string(),
            ffprobe_path: "/usr/bin/ffprobe".to_string(),
            video_codec: VIDEO_CODEC_DEFAULT.to_string(),
        }
    }
}

/// Error related to missing FFmpeg
pub enum FFmpegBadInstallationError {
    FFmpegMissing,
    FFprobeMissing,
}
