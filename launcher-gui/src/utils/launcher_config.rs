// Util to locate the vault configuration file

use std::{
    fs::{self, create_dir_all, read_to_string},
    path::PathBuf,
};

use dirs::config_dir;
use sha2::{Digest, Sha256};

use crate::{
    models::LauncherConfig,
    utils::{file_exists, to_hex_string},
};

fn compute_vault_path_hash_tag(vault_path: &str) -> String {
    let hash = Sha256::digest(vault_path.to_string().into_bytes());
    to_hex_string(&hash)[31..].to_lowercase()
}

fn get_config_file_from_user_config(vault_path: &str) -> Result<String, ()> {
    let mut dir = match config_dir() {
        Some(d) => d,
        None => {
            return Err(());
        }
    };

    dir.push("PersonalMediaVault");
    dir.push("launcher_config");

    if let Err(_) = create_dir_all(&dir) {
        return Err(());
    }

    let vault_path_hash_tag = compute_vault_path_hash_tag(vault_path);

    dir.push(vault_path_hash_tag + ".json");

    match dir.to_str() {
        Some(d) => Ok(d.to_string()),
        None => Err(()),
    }
}

pub fn get_launcher_config_file_general(vault_path: &str) -> String {
    let path_buf: PathBuf = [vault_path, "launcher.config.json"].iter().collect();
    path_buf
        .to_str()
        .expect("Error decoding PathBuf")
        .to_string()
}

/// Resolves the location of the configuration file for the launcher
pub fn get_launcher_config_file(vault_path: &str) -> String {
    let file_general = get_launcher_config_file_general(vault_path);
    let file_specific = match get_config_file_from_user_config(vault_path) {
        Ok(p) => p,
        Err(_) => {
            return file_general;
        }
    };

    if !file_exists(&file_specific) && file_exists(&file_general) {
        let _ = fs::copy(&file_general, &file_specific);
    }

    file_specific
}

pub fn load_launcher_config_from_file(path: &str) -> Result<LauncherConfig, ()> {
    let file_str = match read_to_string(path) {
        Ok(s) => s,
        Err(_) => {
            return Err(());
        }
    };

    let config: LauncherConfig = match serde_json::from_str(&file_str) {
        Ok(c) => c,
        Err(_) => {
            return Err(());
        }
    };

    Ok(config)
}

pub fn write_launcher_to_config_file(path: &str, config: &LauncherConfig) -> Result<(), String> {
    let file_str = match serde_json::to_string(config) {
        Ok(s) => s,
        Err(e) => {
            return Err(e.to_string());
        }
    };

    if let Err(e) = fs::write(path, file_str) {
        return Err(e.to_string());
    }

    Ok(())
}
