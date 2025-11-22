// Callbacks for About component

use crate::{
    constants::{GIT_URL, SITE_URL, VERSION},
    utils::open_url_async,
    MainWindow,
};

pub fn setup_callbacks_about(ui: &MainWindow) {
    ui.set_version(VERSION.into());

    ui.on_visit_website({
        move || {
            open_url_async(SITE_URL);
        }
    });

    ui.on_visit_git({
        move || {
            open_url_async(GIT_URL);
        }
    });
}
