use tonic::{Request, Response, Status};

use crate::grpc::{
    SemanticSearchEngineGrpcServer,
    api::{ClipEmbeddingResponse, ClipTextEmbeddingRequest},
};

pub async fn encode_text(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<ClipTextEmbeddingRequest>,
) -> Result<Response<ClipEmbeddingResponse>, Status> {
    Err(Status::not_found("Not implemented"))
}
