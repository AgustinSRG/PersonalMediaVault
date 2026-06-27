use tonic::{Request, Response, Status};

use crate::{
    NewStoredVector,
    api::{EmptyResponse, InsertVectorsRequest},
    embedding_type_to_int,
    grpc::SemanticSearchEngineGrpcServer,
};

pub async fn insert_vectors(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<InsertVectorsRequest>,
) -> Result<Response<EmptyResponse>, Status> {
    let msg: InsertVectorsRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    for vector_request in &msg.requests {
        if vector_request.features.len() != server.db.get_dimensions() as usize {
            return Err(Status::invalid_argument("Invalid vector size"));
        }
    }

    for vector_request in msg.requests {
        let res = server
            .db
            .insert_vector(NewStoredVector {
                media_id: vector_request.media_id,
                vector_type: embedding_type_to_int(&vector_request.embedding_type()),
                data_hash: vector_request.data_hash,
                embeddings: vector_request.features,
            })
            .await;

        if let Err(err) = res {
            return Err(Status::internal(err.to_string()));
        }
    }

    Ok(Response::new(EmptyResponse {}))
}
