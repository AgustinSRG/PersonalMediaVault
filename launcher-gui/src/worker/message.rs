// Message sent to the worker thread from the event loop

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
    CreateVault,
    CloseVault,
    Finish,
}
