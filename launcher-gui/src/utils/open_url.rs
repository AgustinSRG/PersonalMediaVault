// Utils to open URLs

/// Opens URL in the default browser
pub fn open_url(url: &str) {
    match open::that(url) {
        Ok(()) => println!("Opened '{}' successfully.", url),
        Err(err) => eprintln!("An error occurred when opening '{}': {}", url, err),
    }
}
