use tonic::{Request, Response, Status};

use crate::{
    api::{DeleteVectorsRequest, EmptyResponse},
    grpc::SemanticSearchEngineGrpcServer,
};

pub async fn delete_vectors(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<DeleteVectorsRequest>,
) -> Result<Response<EmptyResponse>, Status> {
    let msg: DeleteVectorsRequest = request.into_inner();

    if !server.auth.check_key(&msg.api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    for vector_id in msg.vector_ids {
        let res = server.db.delete_vector(vector_id).await;

        if let Err(err) = res {
            return Err(Status::internal(err.to_string()));
        }
    }

    Ok(Response::new(EmptyResponse {}))
}
