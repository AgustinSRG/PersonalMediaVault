// Util to locate the vault configuration file

use std::{fs::{self, create_dir_all}, path::PathBuf};

use dirs::config_dir;
use sha2::{Digest, Sha256};

use crate::utils::{file_exists, to_hex_string};

fn compute_vault_path_hash_tag(vault_path: &str) -> String {
    let hash = Sha256::digest(vault_path.to_string().into_bytes());
    to_hex_string(&hash)[31..].to_lowercase()
}

fn get_config_file_from_user_config(vault_path: &str) -> Result<String, ()> {
    let mut dir = match config_dir() {
        Some(d) => d,
        None => {
            return Err(());
        },
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
        None => {
            Err(())
        },
    }
}

pub fn get_launcher_config_file_general(vault_path: &str) -> String {
    let path_buf: PathBuf = [vault_path, "launcher.config.json"].iter().collect();
    path_buf.to_str().expect("Error decoding PathBuf").to_string()
}

/// Resolves the location of the configuration file for the launcher
pub fn get_launcher_config_file(vault_path: &str) -> String {
    let file_general = get_launcher_config_file_general(vault_path);
    let file_specific = match get_config_file_from_user_config(vault_path) {
        Ok(p) => p,
        Err(_) => {
            return file_general;
        },
    };

    if !file_exists(&file_specific) && file_exists(&file_general) {
        let _ = fs::copy(&file_general, &file_specific);
    }

    file_specific
}
