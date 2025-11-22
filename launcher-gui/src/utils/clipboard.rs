// Clipboard utils

use arboard::Clipboard;

use crate::log_debug;

/// Gets the clipboard handle
pub fn get_clipboard() -> Option<Clipboard> {
    match Clipboard::new() {
        Ok(c) => Some(c),
        Err(e) => {
            log_debug!("Clipboard error: {e}");
            None
        }
    }
}

/// Sets the contents of the clipboard
pub fn set_clipboard_contents(clipboard: &mut Option<Clipboard>, contents: &str) {
    if let Some(c) = clipboard {
        if let Err(e) = c.set_text(contents) {
            log_debug!("Clipboard error: {e}");
        }
    }
}
