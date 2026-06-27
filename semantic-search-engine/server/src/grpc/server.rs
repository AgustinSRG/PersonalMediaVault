// GRPC server

use std::net::SocketAddr;

use log::info;
use tokio::net::TcpListener;
use tokio_stream::wrappers::TcpListenerStream;
use tonic::{Request, Response, Status, transport::Server};

use crate::{
    GrpcServerAuth, LoadedClipModel, VectorDatabase,
    api::{
        DeleteVectorsRequest, EmptyResponse, GetVectorsByMediaRequest, InsertVectorsRequest,
        QueryVectorsRequest, VectorListResponse,
    },
    grpc::{
        api::{
            ClipEmbeddingResponse, ClipImageEmbeddingRequest, ClipModelMetadataRequest,
            ClipModelMetadataResponse, ClipTextEmbeddingRequest,
            semantic_search_engine_service_server::{
                SemanticSearchEngineService, SemanticSearchEngineServiceServer,
            },
        },
        methods::{
            delete_vectors, encode_image, encode_text, get_model_metadata, get_vectors_by_media,
            insert_vectors, query_vectors,
        },
    },
};

/// GRPC server
pub struct SemanticSearchEngineGrpcServer {
    /// Model
    pub model: LoadedClipModel,

    /// Auth
    pub auth: GrpcServerAuth,

    /// Vector database
    pub db: VectorDatabase,
}

#[tonic::async_trait]
impl SemanticSearchEngineService for SemanticSearchEngineGrpcServer {
    async fn get_model_metadata(
        &self,
        request: Request<ClipModelMetadataRequest>,
    ) -> Result<Response<ClipModelMetadataResponse>, Status> {
        get_model_metadata(self, request).await
    }

    async fn encode_text(
        &self,
        request: Request<ClipTextEmbeddingRequest>,
    ) -> Result<Response<ClipEmbeddingResponse>, Status> {
        encode_text(self, request).await
    }

    async fn encode_image(
        &self,
        request: Request<tonic::Streaming<ClipImageEmbeddingRequest>>,
    ) -> Result<Response<ClipEmbeddingResponse>, Status> {
        encode_image(self, request).await
    }

    async fn get_vectors_by_media(
        &self,
        request: Request<GetVectorsByMediaRequest>,
    ) -> Result<Response<VectorListResponse>, Status> {
        get_vectors_by_media(self, request).await
    }

    async fn query_vectors(
        &self,
        request: Request<QueryVectorsRequest>,
    ) -> Result<Response<VectorListResponse>, Status> {
        query_vectors(self, request).await
    }

    async fn insert_vectors(
        &self,
        request: Request<InsertVectorsRequest>,
    ) -> Result<Response<EmptyResponse>, Status> {
        insert_vectors(self, request).await
    }

    async fn delete_vectors(
        &self,
        request: Request<DeleteVectorsRequest>,
    ) -> Result<Response<EmptyResponse>, Status> {
        delete_vectors(self, request).await
    }
}

impl SemanticSearchEngineGrpcServer {
    pub fn new(model: LoadedClipModel, auth: GrpcServerAuth, db: VectorDatabase) -> Self {
        Self { model, auth, db }
    }

    pub async fn run(self) -> Result<(), Box<dyn std::error::Error>> {
        let addr: SocketAddr = "127.0.0.0:0".parse()?;

        let listener = TcpListener::bind(addr).await?;

        let local_addr = listener.local_addr()?;

        info!(
            "Starting GRPC Server. Listening on port: {}",
            local_addr.port()
        );

        println!("{}", local_addr.port());

        let incoming = TcpListenerStream::new(listener);

        let mut builder = Server::builder();

        builder
            .add_service(SemanticSearchEngineServiceServer::new(self))
            .serve_with_incoming(incoming)
            .await?;

        Ok(())
    }
}
