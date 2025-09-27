// Utils to open URLs

use std::thread;

/// Opens URL in the default browser
pub fn open_url(url: &str) {
    match open::that(url) {
        Ok(()) => println!("Opened '{}' successfully.", url),
        Err(err) => eprintln!("An error occurred when opening '{}': {}", url, err),
    }
}

/// Opens URL in the default browser (async)
pub fn open_url_async(url: &str) {
    let url_cloned = url.to_string();
    thread::spawn(move || {
        open_url(&url_cloned);
    });
}
