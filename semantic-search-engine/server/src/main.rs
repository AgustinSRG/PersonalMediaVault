// Main

mod database;
mod grpc;
mod model;
mod utils;

use std::{env, path::PathBuf, process::exit, str::FromStr};

use clap::Parser;

use log::{error, info, warn};

pub use database::*;
pub use grpc::*;
pub use model::*;
pub use utils::*;

/// Command line interface
#[derive(Parser)]
#[command(author, version, about, long_about = None)]
struct Cli {
    /// Auto confirm actions
    #[arg(short = 'l', long)]
    pub loglevel: Option<String>,

    /// Max size of the database connection pool
    #[arg(long)]
    pub max_db_pool_size: Option<u32>,

    /// Path to the embeddings model
    pub model_path: PathBuf,

    /// Path to the SQLite database file to store vectors
    pub sqlite_db_path: PathBuf,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    initialize_sqlite_vector_extension();

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

    let model_dimensions = model.get_embed_dim();

    info!(
        "Loaded model from '{}'. Embeddings size: {}",
        cli.model_path.to_string_lossy(),
        model_dimensions,
    );

    let db_passphrase = env::var("SQLITE_CIPHER_PASSPHRASE").unwrap_or("".to_string());

    if db_passphrase.is_empty() {
        error!("Configuration error: Empty value of 'SQLITE_CIPHER_PASSPHRASE'");
        exit(1);
    }

    let max_db_pool_size = cli.max_db_pool_size.unwrap_or(DEFAULT_DB_MAX_POOL_SIZE);

    let db = match VectorDatabase::new(VectorDatabaseConfig {
        path: cli.sqlite_db_path,
        passphrase: db_passphrase,
        vector_dimensions: model_dimensions,
        pool_max_size: max_db_pool_size,
    })
    .await
    {
        Ok(d) => d,
        Err(e) => {
            error!("Could not load vector database: {}", e);
            exit(1);
        }
    };

    let api_key = env::var("API_KEY").unwrap_or("".to_string());

    if api_key.is_empty() {
        warn!("API_KEY is empty. The GRPC server is unprotected.")
    }

    let server = SemanticSearchEngineGrpcServer::new(model, GrpcServerAuth::new(api_key), db);

    if let Err(e) = server.run().await {
        error!("Could not start the server: {}", e);
        exit(1);
    }

    Ok(())
}
