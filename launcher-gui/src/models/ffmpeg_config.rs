// FFmpeg config model

use serde::{Deserialize, Serialize};

pub const H264_CODEC_DEFAULT: &str = "libx264";
pub const H264_CODEC_FREE: &str = "libopenh264";

/// Represents the configuration for the launcher
/// for a given vault path
#[derive(Debug, Serialize, Deserialize)]
pub struct FFmpegConfig {
    #[serde(default, rename = "ffmpeg_path")]
    pub ffmpeg_path: String,

    #[serde(default, rename = "ffprobe_path")]
    pub ffprobe_path: String,

    #[serde(default, rename = "h264_codec")]
    pub h264_codec: String,
}
