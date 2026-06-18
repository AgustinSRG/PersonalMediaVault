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

    let text = msg.text;

    if text.is_empty() {
        return Err(Status::invalid_argument("Text cannot be empty"));
    }

    let clip = server.model.get_clip();

    let embedding = match clip.text.embed_text(&text) {
        Ok(em) => em,
        Err(e) => {
            return Err(Status::internal(format!("Model error: {e}")));
        }
    };

    let features: Vec<f32> = embedding.to_vec();

    Ok(Response::new(ClipEmbeddingResponse { features }))
}
