// Backup progress

use std::time::{Duration, SystemTime};

use slint::Weak;

use crate::{
    log_debug, utils::display_size, MainWindow, VaultBackupCurrentTask, VaultBackupErrorType,
    VaultBackupStatus,
};

pub struct BackupProgressStatus {
    pub bytes_done: u64,
    pub bytes_total: u64,

    pub files_done: u64,
    pub files_total: u64,

    pub indeterminate: bool,

    pub current_task: VaultBackupCurrentTask,

    pub current_file: Option<String>,

    pub file_bytes_done: u64,
    pub file_bytes_total: u64,

    last_update: SystemTime,

    main_window: Weak<MainWindow>,
}

const PROGRESS_UPDATE_INTERVAL: Duration = Duration::from_millis(100);

impl BackupProgressStatus {
    pub fn new(main_window: Weak<MainWindow>) -> Self {
        Self {
            bytes_done: 0,
            bytes_total: 0,
            files_done: 0,
            files_total: 0,
            current_task: VaultBackupCurrentTask::None,
            current_file: None,
            file_bytes_done: 0,
            file_bytes_total: 0,
            last_update: SystemTime::now(),
            main_window,
            indeterminate: true,
        }
    }

    /// Checks if the progress should be updated
    pub fn should_update(&self) -> bool {
        let now = SystemTime::now();

        match now.duration_since(self.last_update) {
            Ok(d) => d >= PROGRESS_UPDATE_INTERVAL,
            Err(err) => {
                log_debug!("SystemTimeError: {err}");
                false
            }
        }
    }

    pub fn start_task(&mut self, task: VaultBackupCurrentTask, indeterminate: bool) {
        self.bytes_done = 0;
        self.files_done = 0;
        self.file_bytes_done = 0;
        self.file_bytes_total = 0;
        self.current_file = None;
        self.current_task = task;
        self.indeterminate = indeterminate;

        self.update();
    }

    pub fn end_task(&mut self) {
        self.files_done = self.files_total;
        self.bytes_done = self.bytes_total;

        self.update();
    }

    pub fn start_file(&mut self, current_file: String, file_bytes_total: u64) {
        self.current_file = Some(current_file);
        self.file_bytes_total = file_bytes_total;
        self.file_bytes_done = 0;
    }

    pub fn end_file(&mut self) {
        self.file_bytes_done = self.file_bytes_total;
    }

    fn get_global_progress(&self) -> (f64, String) {
        if self.indeterminate {
            return (0.0, format!("{}", self.files_done));
        }

        let progress: f64 = if self.bytes_total == 0 {
            if self.files_total == 0 {
                0.0
            } else {
                (self.files_done as f64) / (self.files_total as f64)
            }
        } else {
            (self.bytes_done as f64) / (self.bytes_total as f64)
        };

        let progress_int = (progress * 100.0).round() as i32;

        if self.bytes_total > 0 {
            (
                progress,
                format!(
                    "{}% {} / {} ({} / {})",
                    progress_int,
                    self.files_done,
                    self.files_total,
                    display_size(self.bytes_done),
                    display_size(self.bytes_total)
                ),
            )
        } else if self.files_total > 0 {
            (
                progress,
                format!(
                    "{}% {} / {}",
                    progress_int, self.files_done, self.files_total,
                ),
            )
        } else {
            (1.0, "100% (0 / 0)".to_string())
        }
    }

    fn get_file_progress(&self) -> Option<String> {
        match &self.current_file {
            Some(f) => {
                if self.file_bytes_total != 0 {
                    let progress: f64 =
                        (self.file_bytes_done as f64) / (self.file_bytes_total as f64);

                    let progress_int = (progress * 100.0).round() as i32;

                    Some(format!(
                        "{f} ({}% - {}, {})",
                        progress_int,
                        display_size(self.file_bytes_done),
                        display_size(self.file_bytes_total)
                    ))
                } else {
                    Some(f.to_string())
                }
            }
            None => None,
        }
    }

    pub fn update(&mut self) {
        let indeterminate = self.indeterminate;
        let task = self.current_task;
        let (progress_global, progress_global_str) = self.get_global_progress();
        let file_progress = self.get_file_progress();

        self.last_update = SystemTime::now();

        let wh = self.main_window.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            if win.get_backup_task() != task {
                win.set_backup_task(task);
            }
            win.set_backup_progress(progress_global as f32);
            win.set_backup_progress_global(progress_global_str.into());
            win.set_backup_progress_file(file_progress.unwrap_or("".to_string()).into());
            win.set_backup_progress_indeterminate(indeterminate);
        });
    }

    pub fn set_error(&self, error_type: VaultBackupErrorType, error_details: String) {
        let wh = self.main_window.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_backup_status(VaultBackupStatus::Error);
            win.set_backup_error_type(error_type);
            win.set_backup_error(error_details.into());
        });
    }

    pub fn set_success(&self) {
        let files_str = self.files_total.to_string();
        let bytes_str = display_size(self.bytes_total);
        let wh = self.main_window.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_backup_status(VaultBackupStatus::Success);
            win.set_backup_result_files(files_str.into());
            win.set_backup_result_bytes(bytes_str.into());
        });
    }

    pub fn set_idle(&self) {
        let wh = self.main_window.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_backup_status(VaultBackupStatus::Idle);
        });
    }
}
