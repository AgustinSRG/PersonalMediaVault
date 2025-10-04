// Lock utilities

use std::{fs, path::Path};

use pidlock::Pidlock;

pub const LOCK_FILE_NAME: &'static str = "vault.lock";

/// Tries to lock the vault folder
pub fn lock_vault(path: &str) -> Result<Pidlock, ()> {
    let mut p = Path::new(path).to_path_buf();
    p.push(LOCK_FILE_NAME);

    let mut lock = match Pidlock::new_validated(p) {
        Ok(l) => l,
        Err(e) => {
            eprintln!("Error: {e}");
            return Err(());
        }
    };

    // Try to acquire the lock
    match lock.acquire() {
        Ok(()) => {}
        Err(pidlock::PidlockError::LockExists) => {
            eprintln!("[LOCK] Another instance is already running");
            return Err(());
        }
        Err(e) => {
            eprintln!("[LOCK] Failed to acquire lock: {}", e);
            return Err(());
        }
    }

    Ok(lock)
}

/// Forcefully removes the vault lock
pub fn remove_existing_lock(path: &str) {
     let mut p = Path::new(path).to_path_buf();
    p.push(LOCK_FILE_NAME);
    let _ = fs::remove_file(p);
}
