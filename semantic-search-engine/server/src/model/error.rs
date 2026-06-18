// Error

use std::{error::Error, fmt::Display};

use open_clip_inference::ClipError;

/// Error caused due to a failure loading the model
#[derive(Clone, PartialEq, Eq, Debug)]
pub struct ClipModelLoadError {
    message: String,
}

impl Display for ClipModelLoadError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl Error for ClipModelLoadError {}

impl From<ClipError> for ClipModelLoadError {
    fn from(value: ClipError) -> Self {
        ClipModelLoadError {
            message: value.to_string(),
        }
    }
}
