// Util to find the PMV binaries

use std::path::{absolute, Path};

use crate::utils::{file_exists, folder_exists, get_binary_name, get_dirname};

/// Finds the location of the daemon binary
pub fn find_pmv_daemon_binary() -> Result<String, ()> {
    let mut dir = get_dirname();
    dir.push("bin");
    dir.push(get_binary_name("pmvd"));

    let mut final_path = dir.to_string_lossy().to_string();

    if file_exists(&final_path) {
        return Ok(final_path);
    }

    let p = Path::new("/usr/bin/pmvd").to_string_lossy().to_string();

    if file_exists(&p) {
        return Ok(p);
    }

    dir = get_dirname();
    dir.pop();
    dir.push("backend");
    dir.push(get_binary_name("pmvd"));

    final_path = dir.to_string_lossy().to_string();

    if file_exists(&final_path) {
        return Ok(final_path);
    }

    Err(()) // Not found
}

/// Finds the location of the frontend
pub fn find_pmv_frontend() -> Result<String, ()> {
    let mut dir = get_dirname();
    dir.push("www");

    let mut final_path = dir.to_string_lossy().to_string();

    if folder_exists(&final_path) {
        return Ok(final_path);
    }

    let p = Path::new("/usr/lib/pmv/www").to_string_lossy().to_string();

    if folder_exists(&p) {
        return Ok(p);
    }

    dir = get_dirname();
    dir.pop();
    dir.push("frontend");
    dir.push("dist");

    final_path = dir.to_string_lossy().to_string();

    if folder_exists(&final_path) {
        return Ok(final_path);
    }

    Err(()) // Not found
}
