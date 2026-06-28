// Vector database

use log::debug;
use r2d2::Pool;
use r2d2_sqlite::SqliteConnectionManager;
use rusqlite::params;
use tokio::task;
use zerocopy::IntoBytes;

use crate::{
    NewStoredVector, SqlCipherVecInitializer, StoredVector, VectorDatabaseConfig,
    VectorDatabaseError,
};

const DB_VERSION: &str = "v1";

#[derive(Clone)]
pub struct VectorDatabase {
    pool: Pool<SqliteConnectionManager>,

    dimensions: u32,

    model_integrity: String,
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

        let db = Self {
            pool,
            dimensions: config.vector_dimensions,
            model_integrity: config.model_integrity,
        };

        db.migrate().await?;

        Ok(db)
    }

    fn get_pool(&self) -> Pool<SqliteConnectionManager> {
        self.pool.clone()
    }

    pub fn get_dimensions(&self) -> u32 {
        self.dimensions
    }

    async fn migrate(&self) -> Result<(), VectorDatabaseError> {
        debug!("Migrating database...");

        let config_table_exists = self.check_table_exists("config").await?;

        if !config_table_exists {
            debug!("Config table does not exists. Creating...");
            self.create_config_table().await?;
        }

        let expected_db_version = format!("{}:{}", DB_VERSION, self.model_integrity);

        debug!("Expected database version: {}", expected_db_version);

        let db_version_opt = self.get_db_version().await?;

        match db_version_opt {
            Some(v) => {
                if v != expected_db_version {
                    debug!("Stored version does not match the expected one");

                    self.re_create_vectors_table().await?;

                    debug!("Updating version...");

                    self.update_db_version(expected_db_version).await?;
                } else {
                    debug!("Stored version matches the expected one");
                }
            }
            None => {
                debug!("No database version set");

                self.re_create_vectors_table().await?;

                debug!("Setting version...");

                self.insert_db_version(expected_db_version).await?;
            }
        }

        debug!("Database migration completed");

        Ok(())
    }

    async fn re_create_vectors_table(&self) -> Result<(), VectorDatabaseError> {
        debug!("Re-creating vectors table...");

        self.drop_table("vectors").await?;
        self.create_vectors_table().await?;

        debug!("Re-created vectors table");

        Ok(())
    }

    async fn check_table_exists(&self, name: &'static str) -> Result<bool, VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            let mut stmt =
                conn.prepare("SELECT name FROM sqlite_schema WHERE type='table' AND name=?1")?;

            let tables_names = stmt
                .query_map([name], |row| row.get::<usize, String>(0))?
                .collect::<std::result::Result<Vec<String>, rusqlite::Error>>()?;

            Ok(!tables_names.is_empty())
        });

        handle.await?
    }

    async fn drop_table(&self, name: &'static str) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(&format!("DROP TABLE IF EXISTS {name}"), params![])?;

            Ok(())
        });

        handle.await?
    }

    async fn create_config_table(&self) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(
                "CREATE TABLE config (
                            key TEXT NOT NULL PRIMARY KEY,
                            value TEXT NOT NULL
                        )",
                params![],
            )?;

            Ok(())
        });

        handle.await?
    }

    async fn get_db_version(&self) -> Result<Option<String>, VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            let mut stmt = conn.prepare("SELECT value FROM config WHERE key='version'")?;

            let versions = stmt
                .query_map([], |row| row.get::<usize, String>(0))?
                .collect::<std::result::Result<Vec<String>, rusqlite::Error>>()?;

            if versions.is_empty() {
                return Ok(None);
            }

            let version = versions.first().map_or("", |v| v);

            Ok(Some(version.to_string()))
        });

        handle.await?
    }

    async fn insert_db_version(&self, version: String) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(
                "INSERT INTO config(key, value) VALUES (?, ?)",
                ("version", version),
            )?;

            Ok(())
        });

        handle.await?
    }

    async fn update_db_version(&self, version: String) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(
                "UPDATE config SET value = ?1 WHERE key = ?2",
                (version, "version"),
            )?;

            Ok(())
        });

        handle.await?
    }

    async fn create_vectors_table(&self) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();
        let dimensions = self.dimensions;

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(
                &format!(
                    "CREATE VIRTUAL TABLE vectors USING vec0(
                        id INTEGER PRIMARY KEY,
                        media_id INTEGER NOT NULL,
                        vector_type INTEGER NOT NULL,
                        data_hash TEXT NOT NULL,
                        embedding float[{dimensions}] NOT NULL,
                    )"
                ),
                params![],
            )?;

            conn.execute(
                "CREATE INDEX IF NOT EXISTS vectors_ix_media_id ON vectors (media_id)",
                params![],
            )?;

            conn.execute(
                "CREATE INDEX IF NOT EXISTS vectors_ix_vector_type ON vectors (vector_type)",
                params![],
            )?;

            Ok(())
        });

        handle.await?
    }

    pub async fn insert_vector(&self, vector: NewStoredVector) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute(
                "INSERT INTO vectors(media_id, vector_type, data_hash, embedding) VALUES (?, ?, ?, ?)",
                (vector.media_id as i64, vector.vector_type, &vector.data_hash, vector.embeddings.as_bytes()),
            )?;

            Ok(())
        });

        handle.await?
    }

    pub async fn delete_vector(&self, id: u64) -> Result<(), VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            conn.execute("DELETE FROM vectors WHERE id = ?", params![id as i64])?;

            Ok(())
        });

        handle.await?
    }

    pub async fn get_vectors_by_media_id(
        &self,
        media_id: u64,
    ) -> Result<Vec<StoredVector>, VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            let mut stmt = conn.prepare(
                "SELECT id, media_id, vector_type, data_hash FROM vectors WHERE media_id=?1",
            )?;

            let vectors = stmt
                .query_map([media_id as i64], |row| {
                    Ok(StoredVector {
                        id: row.get::<usize, i64>(0)? as u64,
                        media_id: row.get::<usize, i64>(1)? as u64,
                        vector_type: row.get::<usize, u8>(2)?,
                        data_hash: row.get::<usize, String>(3)?,
                    })
                })?
                .collect::<std::result::Result<Vec<StoredVector>, rusqlite::Error>>()?;

            Ok(vectors)
        });

        handle.await?
    }

    pub async fn query_vectors(
        &self,
        vector: Vec<f32>,
        offset: i64,
        limit: i64,
    ) -> Result<Vec<StoredVector>, VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            let mut stmt = conn.prepare("SELECT id, distance, media_id, vector_type, data_hash FROM vectors WHERE embedding MATCH ?1 ORDER BY distance LIMIT ?2 OFFSET ?3")?;

            let vectors = stmt
                .query_map((vector.as_bytes(), limit, offset), |row| {
                    Ok(StoredVector {
                        id: row.get::<usize, i64>(0)? as u64,
                        media_id: row.get::<usize, i64>(2)? as u64,
                        vector_type: row.get::<usize, u8>(3)?,
                        data_hash: row.get::<usize, String>(4)?,
                    })
                })?
                .collect::<std::result::Result<Vec<StoredVector>, rusqlite::Error>>()?;

            Ok(vectors)
        });

        handle.await?
    }

    pub async fn query_vectors_filtered_by_type(
        &self,
        vector_type: u8,
        vector: Vec<f32>,
        offset: i64,
        limit: i64,
    ) -> Result<Vec<StoredVector>, VectorDatabaseError> {
        let pool = self.get_pool();

        let handle = task::spawn_blocking(move || {
            let conn = pool.get()?;

            let mut stmt = conn.prepare("SELECT id, distance, media_id, vector_type, data_hash FROM vectors WHERE vector_type = ?1 AND embedding MATCH ?2 ORDER BY distance LIMIT ?3 OFFSET ?4")?;

            let vectors = stmt
                .query_map((vector_type, vector.as_bytes(), limit, offset), |row| {
                    Ok(StoredVector {
                        id: row.get::<usize, i64>(0)? as u64,
                        media_id: row.get::<usize, i64>(2)? as u64,
                        vector_type: row.get::<usize, u8>(3)?,
                        data_hash: row.get::<usize, String>(4)?,
                    })
                })?
                .collect::<std::result::Result<Vec<StoredVector>, rusqlite::Error>>()?;

            Ok(vectors)
        });

        handle.await?
    }
}
