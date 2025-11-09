// Task to set the user settings

use slint::Weak;

use crate::{log_debug, worker::WorkerThreadStatus, MainWindow};

pub fn set_user_settings(
    status: &mut WorkerThreadStatus,
    window_handle: &Weak<MainWindow>,
    locale_index: usize,
    theme_index: usize,
) {
    let old_user_settings = status.user_settings.clone();

    status
        .user_settings
        .set_with_indexes(locale_index, theme_index);

    let user_settings = status.user_settings.clone();

    let locale_has_changed = old_user_settings.locale != user_settings.locale;
    let theme_has_changed = old_user_settings.theme != user_settings.theme;

    if !locale_has_changed && !theme_has_changed {
        return;
    }

    if let Err(err) = status.user_settings.save() {
        log_debug!("Error: {err}");
    }

    if theme_has_changed {
        let dark_theme_default = status.dark_theme_default;

        let wh = window_handle.clone();
        let _ = slint::invoke_from_event_loop(move || {
            let win = wh.unwrap();

            if theme_has_changed {
                match user_settings.theme.as_str() {
                    "dark" => {
                        win.invoke_set_dark_theme();
                    }
                    "light" => {
                        win.invoke_set_light_theme();
                    }
                    _ => {
                        if dark_theme_default {
                            win.invoke_set_dark_theme();
                        } else {
                            win.invoke_set_light_theme();
                        }
                    }
                }
            }
        });
    }
}
