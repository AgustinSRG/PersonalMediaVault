// Message sent to the worker thread from the event loop

pub enum LauncherWorkerMessage {
    SelectVaultFolder,
    OpenVault { path: String },
    Finish,
}
