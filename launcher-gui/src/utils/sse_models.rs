// SSE model utils

use std::path::Path;

use crate::utils::{file_exists, folder_exists};

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

/// Validates SSE model
/// Checks if the folder exists and all necessary files are available
/// Returns a boolean (true if valid), and a string indicating the missing file
pub fn validate_sse_model(path: &str) -> (bool, String) {
    if !folder_exists(path) {
        return (false, "".to_string());
    }

    for file in CLIP_MODEL_FILES {
        let mut path_buf = Path::new(path).to_path_buf();
        path_buf.push(file);

        if !file_exists(path_buf) {
            return (false, file.to_string());
        }
    }

    (true, "".to_string())
}
