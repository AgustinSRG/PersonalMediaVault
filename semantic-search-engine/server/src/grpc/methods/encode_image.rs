use std::io::Cursor;

use image::{ImageFormat, ImageReader};
use log::{debug, error};
use tokio_stream::StreamExt;
use tonic::{Request, Response, Status};

use crate::{
    api::clip_image_embedding_request::ImageEmbeddingOneof,
    grpc::{
        SemanticSearchEngineGrpcServer,
        api::{ClipEmbeddingResponse, ClipImageEmbeddingRequest},
    },
};

pub async fn encode_image(
    server: &SemanticSearchEngineGrpcServer,
    request: Request<tonic::Streaming<ClipImageEmbeddingRequest>>,
) -> Result<Response<ClipEmbeddingResponse>, Status> {
    let mut stream = request.into_inner();

    let first_msg = match stream.next().await {
        Some(r) => match r {
            Ok(m) => match m.image_embedding_oneof {
                Some(o) => match o {
                    ImageEmbeddingOneof::Init(c) => c,
                    ImageEmbeddingOneof::Chunk(_) => {
                        return Err(Status::invalid_argument("Missing authentication message"));
                    }
                },
                None => {
                    return Err(Status::invalid_argument("Missing authentication message"));
                }
            },
            Err(_) => {
                return Err(Status::invalid_argument("Missing authentication message"));
            }
        },
        None => {
            return Err(Status::invalid_argument("Missing authentication message"));
        }
    };

    let api_key = first_msg.api_key;

    if !server.auth.check_key(&api_key) {
        return Err(Status::unauthenticated("Wrong API key"));
    }

    let mime_type = first_msg.mime_type;

    let image_format: ImageFormat = match mime_type.as_str() {
        "image/png" => ImageFormat::Png,
        "image/jpg" | "image/jpeg" => ImageFormat::Jpeg,
        "image/gif" => ImageFormat::Gif,
        "image/webp" => ImageFormat::WebP,
        _ => {
            return Err(Status::invalid_argument("Unsupported image type"));
        }
    };

    let mut bytes: Vec<u8> = Vec::new();

    let mut done = false;

    while !done {
        let mut chunk_msg = match stream.next().await {
            Some(r) => match r {
                Ok(m) => match m.image_embedding_oneof {
                    Some(o) => match o {
                        ImageEmbeddingOneof::Init(_) => {
                            return Err(Status::invalid_argument(
                                "Repeated authentication message",
                            ));
                        }
                        ImageEmbeddingOneof::Chunk(c) => c,
                    },
                    None => {
                        continue;
                    }
                },
                Err(err) => {
                    error!("Error receiving chunks: {}", err);
                    return Err(Status::invalid_argument("Error receiving chunks"));
                }
            },
            None => {
                done = true;
                continue;
            }
        };

        bytes.append(&mut chunk_msg.image_chunk);
    }

    if bytes.is_empty() {
        return Err(Status::invalid_argument("Empty image file"));
    }

    let img = match ImageReader::with_format(Cursor::new(bytes), image_format).decode() {
        Ok(i) => i,
        Err(e) => {
            debug!("Error decoding image: {e}");
            return Err(Status::invalid_argument("Invalid image file"));
        }
    };

    debug!(
        "Encoding image: {} [{} x {}]",
        mime_type,
        img.width(),
        img.height()
    );

    let clip = server.model.get_clip();

    let embedding = match clip.vision.embed_image(&img) {
        Ok(em) => em,
        Err(e) => {
            error!("Model error: {e}");
            return Err(Status::internal(format!("Model error: {e}")));
        }
    };

    let features: Vec<f32> = embedding.to_vec();

    debug!(
        "Image: {} [{} x {}], Vector: {:?}",
        mime_type,
        img.width(),
        img.height(),
        features
    );

    Ok(Response::new(ClipEmbeddingResponse { features }))
}
