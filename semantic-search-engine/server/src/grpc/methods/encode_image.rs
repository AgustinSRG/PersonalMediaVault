use tonic::{Request, Response, Status};

use crate::grpc::{
    SemanticSearchEngineGrpcServer,
    api::{ClipEmbeddingResponse, ClipImageEmbeddingRequest},
};

pub async fn encode_image(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<tonic::Streaming<ClipImageEmbeddingRequest>>,
) -> Result<Response<ClipEmbeddingResponse>, Status> {
    Err(Status::not_found("Not implemented"))
}
