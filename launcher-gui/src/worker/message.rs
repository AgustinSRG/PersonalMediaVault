// Message sent to the worker thread from the event loop

use crate::VaultDaemonErrorType;

pub enum LauncherWorkerMessage {
    SelectVaultFolder,
    OpenVault {
        path: String,
    },
    CreateFolderAndOpen,
    ForceOpenVault,
    SetInitialConfig {
        hostname: String,
        port: u16,
        local: bool,
    },
    CreateVault {
        username: String,
        password: String,
    },
    StartVault,
    StopVault,
    VaultStarted {
        daemon_id: u64,
    },
    VaultStartError {
        daemon_id: u64,
        error_type: VaultDaemonErrorType,
        error_details: String,
    },
    VaultStopped {
        daemon_id: u64,
    },
    OpenBrowser,
    OpenLogFile,
    CloseVault,
    Finish,
}
