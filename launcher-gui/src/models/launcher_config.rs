// LauncherConfig model

use serde::{Deserialize, Serialize};

pub const DEFAULT_CACHE_SIZE: i32 = 1024;

#[derive(Serialize, Deserialize, Debug)]
pub struct CacheSize(i32);
impl Default for CacheSize {
    fn default() -> Self {
        CacheSize(DEFAULT_CACHE_SIZE)
    }
}

/// Represents the configuration for the launcher
/// for a given vault path
#[derive(Debug, Serialize, Deserialize)]
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
