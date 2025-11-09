// Locale config model

use std::fs::{self, create_dir_all, read_to_string};

use dirs::config_dir;
use serde::{Deserialize, Serialize};

/// Available locales in the same order the settings ComboBox has them
/// Make sure to update this constant in case new locales are added
pub const AVAILABLE_LOCALES: [&str; 3] = ["", "en", "es"];

/// Available themes in the same order the settings ComboBox has them
pub const AVAILABLE_THEMES: [&str; 3] = ["", "dark", "light"];

/// Represents the user settings
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct UserSettings {
    #[serde(default)]
    pub locale: String,

    #[serde(default)]
    pub theme: String,
}

impl UserSettings {
    pub fn new() -> UserSettings {
        UserSettings {
            locale: "".to_string(),
            theme: "".to_string(),
        }
    }

    pub fn load() -> UserSettings {
        let mut dir = match config_dir() {
            Some(d) => d,
            None => {
                return UserSettings::new();
            }
        };

        dir.push("PersonalMediaVault");

        if create_dir_all(&dir).is_err() {
            return UserSettings::new();
        }

        dir.push("launcher-settings.json");

        let file_str = match read_to_string(dir) {
            Ok(s) => s,
            Err(_) => {
                return UserSettings::new();
            }
        };

        let config: UserSettings = match serde_json::from_str(&file_str) {
            Ok(c) => c,
            Err(_) => {
                return UserSettings::new();
            }
        };

        config
    }

    pub fn save(&self) -> Result<(), String> {
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

        dir.push("launcher-settings.json");

        let file_str = match serde_json::to_string(self) {
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

    pub fn get_locale_index(&self) -> usize {
        AVAILABLE_LOCALES
            .iter()
            .position(|l| l == &self.locale)
            .unwrap_or(0)
    }

    pub fn get_theme_index(&self) -> usize {
        AVAILABLE_THEMES
            .iter()
            .position(|l| l == &self.theme)
            .unwrap_or(0)
    }

    pub fn set_with_indexes(&mut self, locale_index: usize, theme_index: usize) {
        self.locale = AVAILABLE_LOCALES
            .iter()
            .nth(locale_index)
            .unwrap_or(&"")
            .to_string();
        self.theme = AVAILABLE_THEMES
            .iter()
            .nth(theme_index)
            .unwrap_or(&"")
            .to_string();
    }
}
