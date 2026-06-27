// Model integrity computation

use std::path::Path;

use log::debug;
use sha2::{Digest, Sha256};

use crate::{ClipModelLoadError, compute_file_sha256};

const CLIP_MODEL_FILES: [&str; 9] = [
    "model_config.json",
    "open_clip_config.json",
    "tokenizer.json",
    "tokenizer_config.json",
    "special_tokens_map.json",
    "text.onnx",
    "text.onnx.data",
    "visual.onnx",
    "visual.onnx.data",
];

// Computes model integrity by computing the hashes of files
// and creating a compound hash of the model
pub async fn compute_model_integrity(path: &Path) -> Result<String, ClipModelLoadError> {
    let mut hasher = Sha256::new();

    for model_file in CLIP_MODEL_FILES {
        let mut path_buf = path.to_path_buf();
        path_buf.push(model_file);

        debug!("Computing hash of file: {}", path_buf.to_string_lossy());

        match compute_file_sha256(&path_buf).await {
            Ok(file_hash) => {
                hasher.update(file_hash);
            }
            Err(err) => {
                return Err(ClipModelLoadError::new(&format!(
                    "Could not load file {} of model. Error: {}",
                    model_file, err
                )));
            }
        }
    }

    Ok(hex::encode(hasher.finalize()))
}
