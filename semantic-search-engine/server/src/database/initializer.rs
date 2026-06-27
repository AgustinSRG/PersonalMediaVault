// SQlite initializer

use rusqlite::{Connection, ffi::sqlite3_auto_extension};
use sqlite_vec::sqlite3_vec_init;

pub fn initialize_sqlite_vector_extension() {
    unsafe {
        sqlite3_auto_extension(Some(std::mem::transmute(sqlite3_vec_init as *const ())));
    }
}

#[derive(Debug)]
pub struct SqlCipherVecInitializer {
    pub passphrase: String,
}

impl r2d2::CustomizeConnection<Connection, rusqlite::Error> for SqlCipherVecInitializer {
    fn on_acquire(&self, conn: &mut Connection) -> Result<(), rusqlite::Error> {
        // Enforce AES-256 via SQLCipher immediately on connection creation
        conn.pragma_update(None, "key", &self.passphrase)?;

        // Turn on WAL mode to allow multiple concurrent readers while writing
        conn.pragma_update(None, "journal_mode", "WAL")?;

        Ok(())
    }
}
