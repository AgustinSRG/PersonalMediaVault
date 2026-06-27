// Vector database

use r2d2::Pool;
use r2d2_sqlite::SqliteConnectionManager;
use rusqlite::Connection;
use tokio::task;

use crate::{SqlCipherVecInitializer, VectorDatabaseConfig, VectorDatabaseError};

#[derive(Clone)]
pub struct VectorDatabase {
    pool: Pool<SqliteConnectionManager>,
}

impl VectorDatabase {
    /// Creates a new instance of VectorDatabase
    pub async fn new(config: VectorDatabaseConfig) -> Result<VectorDatabase, VectorDatabaseError> {
        let manager = SqliteConnectionManager::file(config.path);

        let pool = Pool::builder()
            .max_size(config.pool_max_size) // Allow up to 10 concurrent connections (readers)
            .connection_customizer(Box::new(SqlCipherVecInitializer {
                passphrase: config.passphrase,
            }))
            .build(manager)?;

        let db = Self { pool };

        db.migrate().await?;

        Ok(db)
    }

    async fn migrate(&self) -> Result<(), VectorDatabaseError> {
        Ok(())
    }
}
