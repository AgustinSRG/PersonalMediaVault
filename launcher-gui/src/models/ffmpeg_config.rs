// FFmpeg config model

use serde::{Deserialize, Serialize};

pub const VIDEO_CODEC_DEFAULT: &str = "libx264";
pub const VIDEO_CODEC_ALTERNATIVE: &str = "libvpx-vp9";

/// Represents the configuration for the launcher
/// for a given vault path
#[derive(Debug, Serialize, Deserialize)]
pub struct FFmpegConfig {
    #[serde(default, rename = "ffmpeg_path")]
    pub ffmpeg_path: String,

    #[serde(default, rename = "ffprobe_path")]
    pub ffprobe_path: String,

    #[serde(default, rename = "video_codec")]
    pub video_codec: String,
}
