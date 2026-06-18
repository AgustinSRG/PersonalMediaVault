use tonic::{Request, Response, Status};

use crate::grpc::{
    SemanticSearchEngineGrpcServer,
    api::{ClipModelMetadataRequest, ClipModelMetadataResponse},
};

pub async fn get_model_metadata(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<ClipModelMetadataRequest>,
) -> Result<Response<ClipModelMetadataResponse>, Status> {
    let msg: ClipModelMetadataRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    Ok(Response::new(ClipModelMetadataResponse {
        embed_dim: server.model.get_embed_dim(),
    }))
}
