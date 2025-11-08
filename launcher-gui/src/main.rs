// Prevent console window in addition to Slint window in Windows release builds when, e.g., starting the app via file manager. Ignored on other platforms.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::{
    env,
    error::Error,
    path::{self, PathBuf},
    sync::mpsc::channel,
};

slint::include_modules!();

mod constants;
mod control;
mod models;
mod utils;
mod worker;

use control::*;
use normalize_path::NormalizePath;

use crate::{utils::{folder_exists, get_dirname}, worker::{LauncherWorkerMessage, run_worker_thread}};

fn main() -> Result<(), Box<dyn Error>> {
    // Load initial data

    if folder_exists("/usr/lib/pmv/lang/") {
        slint::init_translations!("/usr/lib/pmv/lang/");
    } else {
        let mut lang_dir = get_dirname();
        lang_dir.push("lang");

        if folder_exists(&lang_dir) {
            slint::init_translations!(lang_dir);
        } else {
            slint::init_translations!("./lang/");
        }
    }

    // Instantiate the screens

    let main_window = MainWindow::new()?;

    // Create channel for worker

    let (sender, receiver) = channel::<LauncherWorkerMessage>();

    // Setup controllers

    setup_callbacks(&main_window, sender.clone());

    // Create worker

    let worker_join_handle = run_worker_thread(sender.clone(), receiver, main_window.as_weak());

    // Initialization logic

    let args: Vec<String> = env::args().collect();

    if args.len() >= 2 {
        // Open specific vault path
        let p = PathBuf::from(&args[1]);

        if let Ok(abs_path) = path::absolute(p) {
            let abs_path_str = abs_path.normalize().to_string_lossy().to_string();
            main_window.set_launcher_status(LauncherStatus::Opening);
            main_window.set_vault_path(abs_path_str.clone().into());
            let _ = sender.send(LauncherWorkerMessage::OpenVault { path: abs_path_str });
        }
    }

    // Run UI event loop

    main_window.run()?;

    // Done, wait for worker thread to finish

    let _ = sender.send(LauncherWorkerMessage::CloseVault);
    let _ = sender.send(LauncherWorkerMessage::Finish);
    let _ = worker_join_handle.join();

    Ok(())
}
