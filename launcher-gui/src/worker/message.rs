// Message sent to the worker thread from the event loop

use crate::{VaultDaemonErrorType, VaultSelectedTool};

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

    ResetConfig,
    UpdateHostPortConfig {
        host: String,
        port: u16,
        local: bool,
    },
    UpdateTlsConfig {
        enabled: bool,
        cert: String,
        key: String,
    },
    UpdateFFmpegConfig {
        ffmpeg_path: String,
        ffprobe_path: String,
        video_codec: String,
    },
    UpdateOtherConfig {
        cache_size: i32,
        log_requests: bool,
        log_debug: bool,
    },
    SelectFFmpegBinary,
    SelectFFprobeBinary,
    SelectTlsCert,
    SelectTlsKey,

    SelectBackupPath,

    ExportVaultKey {
        username: String,
        password: String,
    },
    CopyToClipboard {
        contents: String,
    },

    RunTool {
        tool: VaultSelectedTool,
    },
    CancelTool,
    ToolSuccess {
        tool_id: u64,
    },
    ToolError {
        tool_id: u64,
        error_details: String,
    },

    CloseVault,
    Finish,
}
