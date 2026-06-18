use tonic::{Request, Response, Status};

use crate::grpc::{
    SemanticSearchEngineGrpcServer,
    api::{ClipEmbeddingResponse, ClipTextEmbeddingRequest},
};

pub async fn encode_text(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<ClipTextEmbeddingRequest>,
) -> Result<Response<ClipEmbeddingResponse>, Status> {
    let msg: ClipTextEmbeddingRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    Err(Status::not_found("Not implemented"))
}
