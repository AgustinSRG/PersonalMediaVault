// Callbacks for user settings popup

use std::sync::mpsc::Sender;

use slint::ComponentHandle;

use crate::{worker::LauncherWorkerMessage, MainWindow};

pub fn setup_callbacks_user_settings(
    ui: &MainWindow,
    worker_sender: Sender<LauncherWorkerMessage>,
) {
    ui.on_apply_user_settings({
        let ui_handle = ui.as_weak();
        let sender = worker_sender.clone();
        move || {
            let ui = ui_handle.unwrap();

            let locale_index = ui.get_settings_locale_index() as usize;
            let theme_index = ui.get_settings_theme_index() as usize;

            let _ = sender.send(LauncherWorkerMessage::SetUserSettings {
                locale_index,
                theme_index,
            });
        }
    });
}
