// Utils to open URLs

use std::thread;

use crate::log_debug;

/// Opens URL in the default browser
pub fn open_url(url: &str) {
    match open::that(url) {
        Ok(()) => {
            log_debug!("Opened '{}' successfully.", url)
        }
        Err(err) => {
            log_debug!("An error occurred when opening '{}': {}", url, err)
        }
    }
}

/// Opens URL in the default browser (async)
pub fn open_url_async(url: &str) {
    let url_cloned = url.to_string();
    thread::spawn(move || {
        open_url(&url_cloned);
    });
}
