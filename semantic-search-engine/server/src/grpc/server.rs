// GRPC server

use std::net::SocketAddr;

use log::info;
use tokio::net::TcpListener;
use tokio_stream::wrappers::TcpListenerStream;
use tonic::{Request, Response, Status, transport::Server};

use crate::{
    LoadedClipModel,
    grpc::{
        api::{
            ClipEmbeddingResponse, ClipImageEmbeddingRequest, ClipModelMetadataRequest,
            ClipModelMetadataResponse, ClipTextEmbeddingRequest,
            semantic_search_engine_service_server::{
                SemanticSearchEngineService, SemanticSearchEngineServiceServer,
            },
        },
        methods::{encode_image, encode_text, get_model_metadata},
    },
};

/// GRPC server
pub struct SemanticSearchEngineGrpcServer {
    /// Model
    pub model: LoadedClipModel,
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
}

impl SemanticSearchEngineGrpcServer {
    pub fn new(model: LoadedClipModel) -> Self {
        Self { model }
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
