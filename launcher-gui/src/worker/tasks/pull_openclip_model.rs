use std::{
    fs::{create_dir_all, File},
    sync::{Arc, Mutex},
    thread,
};

use dirs::config_dir;
use hf_hub::{progress::ProgressHandler, repository::RepoTreeEntry, HFClient};
use slint::Weak;
use tokio_util::sync::CancellationToken;

use crate::{log_debug, utils::file_exists, worker::WorkerThreadStatus, MainWindow};

// Get the common name for the model
fn get_pre_determined_model_name(index: usize) -> &'static str {
    match index {
        1 => "medium", // Medium model
        2 => "big",    // Big model
        _ => "small",  // Small model
    }
}

// Get the HuggingFace name of the model by index
fn get_pre_determined_model_name_hf(index: usize) -> &'static str {
    match index {
        1 => "RuteNL/MobileCLIP2-S3-OpenCLIP-ONNX", // Medium model
        2 => "RuteNL/ViT-SO400M-16-SigLIP2-384-ONNX", // Big model
        _ => "RuteNL/MobileCLIP2-S2-OpenCLIP-ONNX", // Small model
    }
}

struct ModelFileInfo {
    filename: String,
    file_size: u64,
}

struct PreDeterminedOpenClipModel {
    hf_repo_name: String,
    hf_model_name: String,
    location: String,
    complete_file: String,
}

struct ModelDownloadProgressHandler {
    bytes_done: u64,
    total_bytes: u64,
    wh: Weak<MainWindow>,
}

impl ModelDownloadProgressHandler {
    pub fn new(
        bytes_done: u64,
        total_bytes: u64,
        wh: Weak<MainWindow>,
    ) -> ModelDownloadProgressHandler {
        ModelDownloadProgressHandler {
            bytes_done,
            total_bytes,
            wh,
        }
    }
}

impl ProgressHandler for ModelDownloadProgressHandler {
    fn on_progress(&self, event: &hf_hub::progress::ProgressEvent) {
        let wh = self.wh.clone();
        let event_progress: u64 = match event {
            hf_hub::progress::ProgressEvent::Download(
                hf_hub::progress::DownloadEvent::Progress { files },
            ) => files.iter().map(|f| f.bytes_completed).sum(),
            _ => {
                return;
            }
        };
        let progress = if self.total_bytes == 0 {
            0.0
        } else {
            ((self.bytes_done + event_progress) as f32) / (self.total_bytes as f32)
        };

        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_download_progress(progress);
        });
    }
}

fn get_clip_model_details(index: usize) -> Result<PreDeterminedOpenClipModel, String> {
    let name = get_pre_determined_model_name(index);
    let hf_name = get_pre_determined_model_name_hf(index);

    let hf_repo_name = hf_name.split("/").next().unwrap_or("").to_string();
    let hf_model_name = hf_name.split("/").nth(1).unwrap_or("").to_string();

    let mut dir = match config_dir() {
        Some(d) => d,
        None => {
            return Err("Could not locate the model path".to_string());
        }
    };

    dir.push("PersonalMediaVault");
    dir.push("open_clip_models");

    let mut dir_complete_file = dir.clone();

    dir_complete_file.push(name.to_string() + ".complete");

    dir.push(name);

    if let Err(e) = create_dir_all(&dir) {
        return Err(e.to_string());
    }

    Ok(PreDeterminedOpenClipModel {
        hf_repo_name,
        hf_model_name,
        location: dir.to_string_lossy().to_string(),
        complete_file: dir_complete_file.to_string_lossy().to_string(),
    })
}

fn drop_cancellation_token(cancellation_token: Arc<Mutex<Option<CancellationToken>>>) {
    let mut ct = cancellation_token.lock().unwrap();
    *ct = None;
}

async fn download_model_hf(
    window_handle: &Weak<MainWindow>,
    model_details: PreDeterminedOpenClipModel,
    cancellation_token: Arc<Mutex<Option<CancellationToken>>>,
) {
    let client = match HFClient::new() {
        Ok(c) => c,
        Err(e) => {
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_pull_model_error(e.to_string().into());
                win.set_downloading_model(false);
            });
            drop_cancellation_token(cancellation_token);
            return;
        }
    };

    let ct = match cancellation_token.lock() {
        Ok(mg) => match mg.as_ref() {
            Some(t) => t.clone(),
            None => {
                let wh = window_handle.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_pull_model_error("Could not obtain the cancellation token".into());
                    win.set_downloading_model(false);
                });
                return;
            }
        },
        Err(_) => {
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_pull_model_error("Could not obtain the cancellation token".into());
                win.set_downloading_model(false);
            });
            return;
        }
    };

    // Get model info (list of files)

    let model_client = client.model(&model_details.hf_repo_name, &model_details.hf_model_name);

    let info = tokio::select! {
        info_result = model_client.info().send() => {
            match info_result {
                Ok(i) => i,
                Err(e) => {
                    let wh = window_handle.clone();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = wh.unwrap();
                        win.set_pull_model_error(e.to_string().into());
                        win.set_downloading_model(false);
                    });
                    drop_cancellation_token(cancellation_token);
                    return;
                }
            }
        }
        _ = ct.cancelled() => {
            return;
        }
    };

    let siblings = info.siblings.unwrap_or(Vec::new());

    let filenames: Vec<String> = siblings.into_iter().map(|s| s.rfilename).collect();

    let paths_info = tokio::select! {
        p = model_client.get_paths_info().paths(filenames).send() => {
            match p {
                Ok(pi) => pi,
                Err(e) => {
                    let wh = window_handle.clone();
                    let _ = slint::invoke_from_event_loop(move || {
                        let win = wh.unwrap();
                        win.set_pull_model_error(e.to_string().into());
                        win.set_downloading_model(false);
                    });
                    drop_cancellation_token(cancellation_token);
                    return;
                }
            }
        }
        _ = ct.cancelled() => {
            return;
        }
    };

    let mut files: Vec<ModelFileInfo> = Vec::new();
    let mut total_bytes: u64 = 0;

    for entry in paths_info {
        if let RepoTreeEntry::File { path, size, .. } = entry {
            total_bytes += size;
            files.push(ModelFileInfo {
                filename: path,
                file_size: size,
            });
        }
    }

    log_debug!("Total bytes: {}.", total_bytes);

    let mut bytes_done: u64 = 0;

    for file in files {
        if ct.is_cancelled() {
            return;
        }

        log_debug!("Downloading file: {}.", file.filename);
        log_debug!("File size: {}.", file.file_size);

        tokio::select! {
            _ = model_client
                .download_file()
                .filename(file.filename)

                .local_dir(&model_details.location)
                .progress(ModelDownloadProgressHandler::new(
                    bytes_done,
                    total_bytes,
                    window_handle.clone(),
                )).send() => {}
            _ = ct.cancelled() => {
                return;
            }
        };

        bytes_done += file.file_size;

        let progress = if total_bytes == 0 {
            0.0
        } else {
            (bytes_done as f32) / (total_bytes as f32)
        };

        log_debug!("Progress: {}.", progress);

        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_download_progress(progress);
        });
    }

    // Mark as completed

    if let Err(e) = File::create(model_details.complete_file) {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_pull_model_error(e.to_string().into());
            win.set_downloading_model(false);
        });
        drop_cancellation_token(cancellation_token);
        return;
    }

    // Finish

    drop_cancellation_token(cancellation_token);

    let wh = window_handle.clone();
    let _ = slint::invoke_from_event_loop(move || {
        let win = wh.unwrap();
        win.set_downloading_model(false);
        win.set_model_path(model_details.location.into());
        win.set_dirty_sse(true);
        win.invoke_close_popup_clip_model();
    });
}

pub fn pull_pre_determined_open_clip_model(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    size_index: usize,
) {
    let model_details = match get_clip_model_details(size_index) {
        Ok(d) => d,
        Err(e) => {
            let wh = window_handle.clone();
            let _ = slint::invoke_from_event_loop(move || {
                let win = wh.unwrap();
                win.set_pull_model_error(e.into());
                win.set_busy(false);
            });
            return;
        }
    };

    // Check if model is already pulled

    let file_exists = file_exists(&model_details.complete_file);

    if file_exists {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_model_path(model_details.location.into());
            win.set_dirty_sse(true);
            win.set_busy(false);
            win.invoke_close_popup_clip_model();
        });
        return;
    }

    // Download from HuggingFace

    {
        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();
            win.set_busy(false);
            win.set_downloading_model(true);
            win.set_download_progress(0.0);
        });
    }

    status.cancel_model_pull();

    let cancellation_token = status.make_pull_model_cancellation_token();

    let window_handle_cloned = window_handle.clone();

    thread::spawn(move || {
        let rt = match tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()
        {
            Ok(r) => r,
            Err(e) => {
                let wh = window_handle_cloned.clone();
                let _ = slint::invoke_from_event_loop(move || {
                    let win = wh.unwrap();
                    win.set_pull_model_error(e.to_string().into());
                    win.set_downloading_model(false);
                });
                return;
            }
        };

        rt.block_on(async move {
            download_model_hf(&window_handle_cloned, model_details, cancellation_token).await;
        })
    });
}
