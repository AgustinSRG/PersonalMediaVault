use tonic::{Request, Response, Status};

use crate::{
    api::{QueryVectorsRequest, VectorListItem, VectorListResponse},
    grpc::SemanticSearchEngineGrpcServer,
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

    let res = match server
        .db
        .query_vectors(msg.features, msg.continuation_token, limit)
        .await
    {
        Ok(l) => l,
        Err(err) => {
            return Err(Status::internal(err.to_string()));
        }
    };

    let continuation_token = res.last().map(|v| v.distance);

    Ok(Response::new(VectorListResponse {
        vectors: res
            .iter()
            .map(|v| VectorListItem {
                vector_id: v.id,
                media_id: v.media_id,
                data_hash: v.data_hash.clone(),
            })
            .collect(),
        continuation_token,
    }))
}
