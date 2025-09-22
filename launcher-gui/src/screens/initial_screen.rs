// Initial screen

use std::error::Error;

use slint::ComponentHandle;

use crate::{
    constants::{GIT_URL, SITE_URL, VERSION},
    utils::open_url,
    InitialScreen,
};

pub fn show_initial_screen() -> Result<(), Box<dyn Error>> {
    let ui = InitialScreen::new()?;

    ui.set_version(VERSION.into());

    ui.on_open_vault_folder({
        move || {
            let dialog = rfd::FileDialog::new();
            match dialog.pick_folder() {
                Some(folder_path) => match folder_path.to_str() {
                    Some(p) => {
                        println!("Open vault: {p}");
                    }
                    None => {}
                },
                None => {}
            }
        }
    });

    ui.on_open_default_vault({
        // TODO
        move || {
            println!("Pressed the button to open the default vault!");
        }
    });

    ui.on_visit_website({
        move || {
            open_url(SITE_URL);
        }
    });

    ui.on_visit_git({
        move || {
            open_url(GIT_URL);
        }
    });

    ui.show()?;

    Ok(())
}
