// Stored vectors

/// Stored vector (for lists)
pub struct StoredVector {
    /// Vector ID
    pub id: u64,

    /// Media ID
    pub media_id: u64,

    /// Vector type
    pub vector_type: u8,

    /// Data hash
    pub data_hash: String,
}

/// New stored vector
pub struct NewStoredVector {
    /// Media ID
    pub media_id: u64,

    /// Vector type
    pub vector_type: u8,

    /// Data hash
    pub data_hash: String,

    /// Embeddings
    pub embeddings: Vec<f32>,
}
