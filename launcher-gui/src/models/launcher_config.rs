// LauncherConfig model

use serde::{Deserialize, Serialize};

pub const DEFAULT_CACHE_SIZE: i32 = 1024;

#[derive(Serialize, Deserialize, Debug, Copy, Clone)]
pub struct CacheSize(i32);
impl Default for CacheSize {
    fn default() -> Self {
        CacheSize(DEFAULT_CACHE_SIZE)
    }
}

impl CacheSize {
    pub fn new(i: i32) -> Self {
        CacheSize(i)
    }

    pub fn as_i32(&self) -> i32 {
        self.0
    }
}

/// Represents the configuration for the launcher
/// for a given vault path
#[derive(Debug, Serialize, Deserialize, Default, Clone)]
pub struct LauncherConfig {
    #[serde(default)]
    pub path: String,

    #[serde(default)]
    pub hostname: String,

    #[serde(default)]
    pub port: u16,

    #[serde(default)]
    pub local: bool,

    #[serde(default, rename = "ssl_cert")]
    pub ssl_cert: String,

    #[serde(default, rename = "ssl_key")]
    pub ssl_key: String,

    #[serde(default, rename = "cache_size")]
    pub cache_size: CacheSize,

    #[serde(default, rename = "log_requests")]
    pub log_requests: bool,

    #[serde(default)]
    pub debug: bool,
}

impl LauncherConfig {
    pub fn new() -> LauncherConfig {
        LauncherConfig {
            ..Default::default()
        }
    }

    pub fn has_ssl(&self) -> bool {
        !self.ssl_cert.is_empty() && !self.ssl_key.is_empty()
    }

    pub fn get_health_check_url(&self, launcher_tag: &str) -> String {
        let protocol = if self.has_ssl() { "https:" } else { "http:" };

        let port = self.port;

        format!("{protocol}//localhost:{port}/api/admin/launcher/{launcher_tag}")
    }

    pub fn get_browser_url(&self) -> String {
        let protocol = if self.has_ssl() { "https:" } else { "http:" };

        let hostname = if self.hostname.is_empty() {
            "localhost".to_string()
        } else {
            self.hostname.clone()
        };

        let port = self.port;

        let port_part = if self.has_ssl() {
            if port != 443 {
                format!(":{port}")
            } else {
                "".to_string()
            }
        } else if port != 80 {
            format!(":{port}")
        } else {
            "".to_string()
        };

        format!("{protocol}//{hostname}{port_part}")
    }
}
