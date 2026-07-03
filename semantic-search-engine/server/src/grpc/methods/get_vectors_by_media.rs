use tonic::{Request, Response, Status};

use crate::{
    api::{GetVectorsByMediaRequest, VectorListItem, VectorListResponse},
    grpc::SemanticSearchEngineGrpcServer,
};

pub async fn get_vectors_by_media(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<GetVectorsByMediaRequest>,
) -> Result<Response<VectorListResponse>, Status> {
    let msg: GetVectorsByMediaRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    let res = match server.db.get_vectors_by_media_id(msg.media_id).await {
        Ok(l) => l,
        Err(err) => {
            return Err(Status::internal(err.to_string()));
        }
    };

    Ok(Response::new(VectorListResponse {
        vectors: res
            .iter()
            .map(|v| VectorListItem {
                vector_id: v.id,
                media_id: v.media_id,
                data_hash: v.data_hash.clone(),
            })
            .collect(),
        continuation_token: None,
    }))
}
