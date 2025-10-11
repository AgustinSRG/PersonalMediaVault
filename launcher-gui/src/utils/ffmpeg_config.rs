// FFmpeg config utils

use std::{
    collections::{HashMap, HashSet},
    env,
    fs::{self, create_dir_all, read_to_string},
    path::Path,
    process::Command,
};

use dirs::config_dir;

use crate::{
    log_debug,
    models::{
        FFmpegBadInstallationError, FFmpegConfig, VIDEO_CODEC_ALTERNATIVE, VIDEO_CODEC_DEFAULT,
    },
    utils::{file_exists, get_binary_name, get_dirname},
};

pub fn load_ffmpeg_config_from_file() -> Result<FFmpegConfig, ()> {
    let mut dir = match config_dir() {
        Some(d) => d,
        None => {
            return Err(());
        }
    };

    dir.push("PersonalMediaVault");

    if let Err(_) = create_dir_all(&dir) {
        return Err(());
    }

    dir.push("ffmpeg.json");

    let file_str = match read_to_string(dir) {
        Ok(s) => s,
        Err(_) => {
            return Err(());
        }
    };

    let config: FFmpegConfig = match serde_json::from_str(&file_str) {
        Ok(c) => c,
        Err(_) => {
            return Err(());
        }
    };

    Ok(config)
}

pub fn write_ffmpeg_to_config_file(config: &FFmpegConfig) -> Result<(), String> {
    let mut dir = match config_dir() {
        Some(d) => d,
        None => {
            return Err("Could not resolve the user configuration folder".to_string());
        }
    };

    dir.push("PersonalMediaVault");

    if let Err(e) = create_dir_all(&dir) {
        return Err(e.to_string());
    }

    dir.push("ffmpeg.json");

    let file_str = match serde_json::to_string(config) {
        Ok(s) => s,
        Err(e) => {
            return Err(e.to_string());
        }
    };

    if let Err(e) = fs::write(dir, file_str) {
        return Err(e.to_string());
    }

    Ok(())
}

fn detect_video_codecs(config: &FFmpegConfig) -> Result<HashSet<String>, String> {
    let mut cmd = Command::new(&config.ffmpeg_path);

    cmd.arg("-encoders");

    let out = match cmd.output() {
        Ok(o) => o,
        Err(e) => {
            return Err(e.to_string());
        }
    };

    let out_str = String::from_utf8_lossy(&out.stdout).to_string();

    let mut available_codecs: HashSet<String> = HashSet::new();

    let lines = out_str.split("\n");

    for line in lines {
        let line_parts: Vec<&str> = line.trim().split(" ").collect();

        if line_parts.len() < 2 {
            continue;
        }

        let codec_name = line_parts[1];

        available_codecs.insert(codec_name.to_string());
    }

    Ok(available_codecs)
}

fn detect_video_codec(config: &FFmpegConfig) -> Result<&'static str, String> {
    let available_codecs = detect_video_codecs(config)?;

    if available_codecs.contains(VIDEO_CODEC_DEFAULT) {
        return Ok(VIDEO_CODEC_DEFAULT);
    }

    if available_codecs.contains(VIDEO_CODEC_ALTERNATIVE) {
        return Ok(VIDEO_CODEC_ALTERNATIVE);
    }

    Err("Unavailable video codec".to_string())
}

/// Checks if a codec is in the list of available codecs
pub fn check_ffmpeg_codec(config: &FFmpegConfig) -> bool {
    match detect_video_codecs(config) {
        Ok(available_codecs) => available_codecs.contains(&config.video_codec),
        Err(_) => false,
    }
}

pub fn load_ffmpeg_config() -> Result<FFmpegConfig, FFmpegBadInstallationError> {
    let mut result = FFmpegConfig {
        ffmpeg_path: "".to_string(),
        ffprobe_path: "".to_string(),
        video_codec: "".to_string(),
    };

    if let Ok(c) = load_ffmpeg_config_from_file() {
        result = c;
    }

    if result.ffmpeg_path.is_empty() || !file_exists(&result.ffmpeg_path) {
        let mut dir = get_dirname();
        dir.push("bin");
        dir.push(get_binary_name("ffmpeg"));
        result.ffmpeg_path = dir.to_string_lossy().to_string();

        if !file_exists(&result.ffmpeg_path) {
            if env::consts::OS == "windows" {
                let p = Path::new("C:\\ffmpeg\\bin\\ffmpeg.exe");
                result.ffmpeg_path = p.to_string_lossy().to_string();
            } else {
                let p = Path::new("/usr/bin/ffmpeg");
                result.ffmpeg_path = p.to_string_lossy().to_string();
            }

            if !file_exists(&result.ffmpeg_path) {
                return Err(FFmpegBadInstallationError::FFmpegMissing);
            }
        }
    }

    if result.ffprobe_path.is_empty() || !file_exists(&result.ffprobe_path) {
        let mut dir = get_dirname();
        dir.push("bin");
        dir.push(get_binary_name("ffprobe"));
        result.ffprobe_path = dir.to_string_lossy().to_string();

        if !file_exists(&result.ffprobe_path) {
            if env::consts::OS == "windows" {
                let p = Path::new("C:\\ffmpeg\\bin\\ffprobe.exe");
                result.ffprobe_path = p.to_string_lossy().to_string();
            } else {
                let p = Path::new("/usr/bin/ffprobe");
                result.ffprobe_path = p.to_string_lossy().to_string();
            }

            if !file_exists(&result.ffprobe_path) {
                return Err(FFmpegBadInstallationError::FFprobeMissing);
            }
        }
    }

    if result.video_codec.is_empty() {
        result.video_codec = match detect_video_codec(&result) {
            Ok(c) => c.to_string(),
            Err(e) => {
                log_debug!("[WARNING] Could not detect a video encoder in your FFMpeg installation. The will lead to errors when trying to encode videos. DEtails: {e}");

                VIDEO_CODEC_DEFAULT.to_string()
            }
        }
    }

    Ok(result)
}
