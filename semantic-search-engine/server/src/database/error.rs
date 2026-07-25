// Errors

use std::{error::Error, fmt::Display};

use tokio::task::JoinError;

/// Error caused due to a failure loading the model
#[derive(Clone, PartialEq, Eq, Debug)]
pub struct VectorDatabaseError {
    message: String,
}

impl Display for VectorDatabaseError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.message)
    }
}

impl Error for VectorDatabaseError {}

impl From<r2d2::Error> for VectorDatabaseError {
    fn from(value: r2d2::Error) -> Self {
        VectorDatabaseError {
            message: value.to_string(),
        }
    }
}

impl From<rusqlite::Error> for VectorDatabaseError {
    fn from(value: rusqlite::Error) -> Self {
        VectorDatabaseError {
            message: value.to_string(),
        }
    }
}

impl From<JoinError> for VectorDatabaseError {
    fn from(value: JoinError) -> Self {
        VectorDatabaseError {
            message: value.to_string(),
        }
    }
}
