// Close vault logic

use crate::worker::WorkerThreadStatus;

pub fn close_vault(status: &mut WorkerThreadStatus) {
    // TODO: Kill the daemon process

    // Remove the lock of the folder
    status.vault_lock = None;
}
