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

use crate::{
    models::UserSettings,
    worker::{run_worker_thread, LauncherWorkerMessage},
};

fn main() -> Result<(), Box<dyn Error>> {
    // Instantiate the screens

    let main_window = MainWindow::new()?;

    // Create channel for worker

    let (sender, receiver) = channel::<LauncherWorkerMessage>();

    // Setup controllers

    setup_callbacks(&main_window, sender.clone());

    // Initial user settings

    let dark_theme_default = main_window.get_is_dark_theme_default();

    let user_settings = UserSettings::load();

    main_window.set_settings_locale_index(user_settings.get_locale_index() as i32);
    main_window.set_settings_theme_index(user_settings.get_theme_index() as i32);

    if !user_settings.locale.is_empty() {
        let _ = slint::select_bundled_translation(&user_settings.locale);
    }

    if !user_settings.theme.is_empty() {
        if user_settings.theme == "dark" {
            main_window.invoke_set_dark_theme();
        } else {
            main_window.invoke_set_light_theme();
        }
    }

    // Create worker

    let worker_join_handle = run_worker_thread(
        sender.clone(),
        receiver,
        main_window.as_weak(),
        user_settings,
        dark_theme_default,
    );

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
