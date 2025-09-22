// Prevent console window in addition to Slint window in Windows release builds when, e.g., starting the app via file manager. Ignored on other platforms.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::error::Error;

use crate::screens::show_initial_screen;

slint::include_modules!();

mod screens;
mod status;
mod constants;
mod utils;

fn main() -> Result<(), Box<dyn Error>> {
    slint::init_translations!(concat!(env!("CARGO_MANIFEST_DIR"), "/lang/"));

    show_initial_screen()?;

    slint::run_event_loop()?;

    Ok(())
}
