#[cfg(unix)]
use std::process::Command;

/// Prevents creating a window when creating the process
#[cfg(windows)]
pub fn command_no_window(cmd: &mut Command) -> std::io::Result<()> {
    cmd.creation_flags(CREATE_NO_WINDOW);
    Ok(())
}

#[cfg(unix)]
pub fn command_no_window(_cmd: &mut Command) -> std::io::Result<()> {
    // This is only a Windows problem
    Ok(())
}
