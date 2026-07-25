// Stored vectors

/// Stored vector (for lists)
pub struct StoredVector {
    /// Vector ID
    pub id: u64,

    /// Media ID
    pub media_id: u64,

    /// Data hash
    pub data_hash: String,
}

pub struct StoredVectorWithDistance {
    /// Vector ID
    pub id: u64,

    /// Media ID
    pub media_id: u64,

    /// Data hash
    pub data_hash: String,

    // Distance
    pub distance: f32,
}

/// New stored vector
pub struct NewStoredVector {
    /// Media ID
    pub media_id: u64,

    /// Data hash
    pub data_hash: String,

    /// Embeddings
    pub embeddings: Vec<f32>,
}
