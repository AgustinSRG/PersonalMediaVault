use tonic::{Request, Response, Status};

use crate::{
    api::{EmbeddingType, QueryVectorsRequest, VectorListItem, VectorListResponse},
    embedding_type_to_int,
    grpc::SemanticSearchEngineGrpcServer,
    int_to_embedding_type,
};

pub async fn query_vectors(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<QueryVectorsRequest>,
) -> Result<Response<VectorListResponse>, Status> {
    let msg: QueryVectorsRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    if msg.features.len() != server.db.get_dimensions() as usize {
        return Err(Status::invalid_argument("Invalid vector size"));
    }

    let limit = msg.limit as i64;

    let offset = msg.offset.unwrap_or(0) as i64;

    let res = match msg.embedding_type {
        Some(embedding_type) => match server
            .db
            .query_vectors_filtered_by_type(
                embedding_type_to_int(
                    &EmbeddingType::try_from(embedding_type).unwrap_or(EmbeddingType::Text),
                ),
                msg.features,
                offset,
                limit,
            )
            .await
        {
            Ok(l) => l,
            Err(err) => {
                return Err(Status::internal(err.to_string()));
            }
        },
        None => match server.db.query_vectors(msg.features, offset, limit).await {
            Ok(l) => l,
            Err(err) => {
                return Err(Status::internal(err.to_string()));
            }
        },
    };

    Ok(Response::new(VectorListResponse {
        vectors: res
            .iter()
            .map(|v| VectorListItem {
                vector_id: v.id,
                media_id: v.media_id,
                embedding_type: int_to_embedding_type(v.vector_type).into(),
                data_hash: v.data_hash.clone(),
            })
            .collect(),
    }))
}
