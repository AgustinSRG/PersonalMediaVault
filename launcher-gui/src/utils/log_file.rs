// Log files utils

use std::{
    fs::{create_dir_all, read_dir, remove_file},
    path::{Path, PathBuf},
    process,
};

use dirs::cache_dir;

use crate::utils::folder_exists;
use chrono::Datelike;

const LIMIT_LOG_FILES: usize = 100;

pub fn get_logs_path() -> Result<String, String> {
    let mut dir = match cache_dir() {
        Some(d) => d,
        None => {
            return Err("Could not find a path to store the log file".to_string());
        }
    };

    dir.push("PersonalMediaVault");
    dir.push("logs");

    if !folder_exists(&dir) {
        if let Err(e) = create_dir_all(&dir) {
            return Err(e.to_string());
        }
    }

    Ok(dir.to_string_lossy().to_string())
}

pub fn get_log_file() -> Result<String, String> {
    let logs_folder = get_logs_path()?;

    // Remove old log files if there are too many

    let mut log_files: Vec<String> = match read_dir(&logs_folder) {
        Ok(l) => l
            .map(|l| match l {
                Ok(de) => de.file_name().to_string_lossy().to_string(),
                Err(_) => "".to_string(),
            })
            .filter(|l| l.ends_with(".log"))
            .collect(),
        Err(_) => todo!(),
    };

    log_files.sort();

    if log_files.len() > LIMIT_LOG_FILES {
        for file in log_files.iter().take(log_files.len() - LIMIT_LOG_FILES) {
            let mut pb: PathBuf = Path::new(&logs_folder).to_path_buf();
            pb.push(file);
            let _ = remove_file(pb);
        }
    }

    let pid = process::id();

    let current_date = chrono::Utc::now();

    let ts = current_date.timestamp_millis();

    let year = current_date.year().to_string();
    let mut month = current_date.month().to_string();
    if month.len() < 2 {
        month = "0".to_string() + &month;
    }
    let mut day: String = current_date.day().to_string();
    if day.len() < 2 {
        day = "0".to_string() + &day;
    }

    let file_name = format!("{year}-{month}-{day}-{ts}-{pid}.log");

    let mut pb: PathBuf = Path::new(&logs_folder).to_path_buf();
    pb.push(file_name);

    Ok(pb.to_string_lossy().to_string())
}
