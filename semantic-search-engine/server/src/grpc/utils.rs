// Utils

use crate::api::EmbeddingType;

pub fn embedding_type_to_int(t: &EmbeddingType) -> u8 {
    match t {
        EmbeddingType::Text => 0,
        EmbeddingType::Image => 1,
    }
}

pub fn int_to_embedding_type(i: u8) -> EmbeddingType {
    match i {
        1 => EmbeddingType::Image,
        _ => EmbeddingType::Text,
    }
}
