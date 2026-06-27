// Database configuration

use std::path::PathBuf;

/// Default max pool size
pub const DEFAULT_DB_MAX_POOL_SIZE: u32 = 10;

/// Vector database config
pub struct VectorDatabaseConfig {
    // Path to the database file
    pub path: PathBuf,

    // Passphrase to encrypt the database
    pub passphrase: String,

    // Dimensions of vectors
    pub vector_dimensions: u32,

    // Max size of the database pool
    pub pool_max_size: u32,
}
