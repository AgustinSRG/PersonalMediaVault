// Close vault logic

use crate::worker::WorkerThreadStatus;

pub fn close_vault(status: &mut WorkerThreadStatus) {
    // Stop vault
    if let Some(p) = &status.daemon_process {
        let _ = p.kill();
    }

    if let Some(r) = &status.daemon_process_wait_receiver {
        let _ = r.recv();
        status.daemon_process_wait_receiver = None;
    }

    // Remove the lock of the folder
    status.vault_lock = None;
}
