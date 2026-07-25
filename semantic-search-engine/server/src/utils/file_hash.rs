use sha2::{Digest, Sha256};
use std::path::Path;
use tokio::fs::File;
use tokio::io::{self, AsyncReadExt};

/// Computes the SHA-256 hash of a file asynchronously using a buffer.
pub async fn compute_file_sha256<P: AsRef<Path>>(path: P) -> io::Result<Vec<u8>> {
    let mut file = File::open(path).await?;

    let mut hasher = Sha256::new();

    let mut buffer = [0u8; 65536];

    loop {
        let bytes_read = file.read(&mut buffer).await?;

        if bytes_read == 0 {
            break;
        }

        hasher.update(&buffer[..bytes_read]);
    }

    Ok(hasher.finalize().to_vec())
}
