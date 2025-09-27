// Background worker thread

use std::{
    sync::mpsc::{Receiver, Sender},
    thread::{spawn, JoinHandle},
};

mod message;
pub use message::*;

mod tasks;
use slint::Weak;
pub use tasks::*;

use crate::MainWindow;

/// Run the launcher main worker thread
pub fn run_worker_thread(
    sender: Sender<LauncherWorkerMessage>,
    receiver: Receiver<LauncherWorkerMessage>,
    window_handle: Weak<MainWindow>,
) -> JoinHandle<()> {
    // Spawn thread
    spawn(move || loop {
        match receiver.recv() {
            Ok(msg) => match msg {
                LauncherWorkerMessage::OpenVault { path } => {
                    println!("Open vault: {path}");
                }
                LauncherWorkerMessage::SelectVaultFolder => {
                    select_vault_folder(sender.clone(), window_handle.clone());
                }
                LauncherWorkerMessage::Finish => {
                    return;
                }
            },
            Err(err) => {
                eprintln!("Error: {}", err);
                return;
            }
        }
    })
}
