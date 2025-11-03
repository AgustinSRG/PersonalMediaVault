// Utility to resolve the default vault path

use std::path::{absolute, PathBuf};

use dirs::config_dir;
use normalize_path::NormalizePath;

/// Resolves the default vault path based on the OS
pub fn get_default_vault_path() -> String {
    let dir = config_dir();

    match dir {
        Some(user_config_dir) => {
            let mut vault_path = user_config_dir.clone();

            vault_path.push("PersonalMediaVault");
            vault_path.push("vault");

            let abs = absolute(vault_path);
            match abs {
                Ok(p) => p.normalize().to_string_lossy().to_string(),
                Err(_) => {
                    panic!("Could not resolve Path");
                }
            }
        }
        None => {
            let path: PathBuf = ["PersonalMediaVault", "vault"].iter().collect();
            let abs = absolute(path);
            match abs {
                Ok(p) => p.normalize().to_string_lossy().to_string(),
                Err(_) => {
                    panic!("Could not resolve Path");
                }
            }
        }
    }
}
