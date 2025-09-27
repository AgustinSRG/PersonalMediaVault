// Prevent console window in addition to Slint window in Windows release builds when, e.g., starting the app via file manager. Ignored on other platforms.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::{error::Error, sync::mpsc::channel};

slint::include_modules!();

mod control;
mod status;
mod constants;
mod utils;
mod worker;

use control::*;

use crate::worker::{run_worker_thread, LauncherWorkerMessage};

fn main() -> Result<(), Box<dyn Error>> {
    // Load initial data

    slint::init_translations!(concat!(env!("CARGO_MANIFEST_DIR"), "/lang/"));

    // Instantiate the screens

    let main_window = MainWindow::new()?;

    // Create channel for worker

    let (sender, receiver) = channel::<LauncherWorkerMessage>();

    // Setup controllers

    setup_callbacks(&main_window, sender.clone());

    // Create worker
    
    let worker_join_handle = run_worker_thread(sender.clone(), receiver, main_window.as_weak());

    // Run UI event loop

    main_window.run()?;

    // Done, wait for worker thread to finish

    let _ = sender.send(LauncherWorkerMessage::Finish);
    let _ = worker_join_handle.join();

    Ok(())
}
