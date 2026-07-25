use log::{debug, error};
use tonic::{Request, Response, Status};

use crate::{
    api::GetVectorEmbeddingsRequest,
    grpc::{SemanticSearchEngineGrpcServer, api::ClipEmbeddingResponse},
};

pub async fn get_vector_embeddings(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<GetVectorEmbeddingsRequest>,
) -> Result<Response<ClipEmbeddingResponse>, Status> {
    let msg: GetVectorEmbeddingsRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    let embedding = match server.db.get_vector_embeddings(msg.vector_id).await {
        Ok(embedding_opt) => match embedding_opt {
            Some(em) => em,
            None => {
                return Err(Status::not_found("Vector not found in the database"));
            }
        },
        Err(e) => {
            error!("DB error: {e}");
            return Err(Status::internal(format!("DB error: {e}")));
        }
    };
    let features: Vec<f32> = embedding.to_vec();

    debug!("ID: {}, Vector size: {}", msg.vector_id, features.len());

    Ok(Response::new(ClipEmbeddingResponse { features }))
}
