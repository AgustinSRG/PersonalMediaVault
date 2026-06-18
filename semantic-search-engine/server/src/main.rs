// Main

mod grpc;
mod model;

use std::{path::PathBuf, process::exit, str::FromStr};

use clap::Parser;

use log::{error, info, warn};

pub use grpc::*;
pub use model::*;

/// Command line interface
#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    /// Auto confirm actions
    #[arg(short = 'l', long)]
    pub loglevel: Option<String>,

    pub model_path: PathBuf,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let cli = Cli::parse();

    if let Some(loglevel) = &cli.loglevel {
        let log_level: log::Level = log::Level::from_str(loglevel).unwrap_or_else(|err| {
            warn!("Error parsing log level: {}. Using INFO as default.", err);

            log::Level::Info
        });
        _ = simple_logger::init_with_level(log_level);
    } else {
        _ = simple_logger::init_with_level(log::Level::Info);
    }

    let model = match LoadedClipModel::load(&cli.model_path) {
        Ok(m) => m,
        Err(e) => {
            error!(
                "Could not load model from {}: {}",
                cli.model_path.to_string_lossy(),
                e
            );
            exit(1);
        }
    };

    info!(
        "Loaded model from '{}'. Embeddings size: {}",
        cli.model_path.to_string_lossy(),
        model.get_embed_dim()
    );

    let server = SemanticSearchEngineGrpcServer::new(model);

    if let Err(e) = server.run().await {
        error!("Could not start the server: {}", e);
        exit(1);
    }

    Ok(())
}
