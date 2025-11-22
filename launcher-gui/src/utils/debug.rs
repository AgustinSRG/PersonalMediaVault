// Debug messages

#[macro_export]
macro_rules! log_debug {
    // This marco logs an ERROR message, only if the ERROR level is enabled
    // The first argument must be the logger
    // The second argument must be the message to log, as a string
    ($($arg:tt)*) => {
        if cfg!(debug_assertions) {
            eprintln!($($arg)*);
        }
    };
}
