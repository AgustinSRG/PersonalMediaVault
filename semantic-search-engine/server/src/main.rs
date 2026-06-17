// Main

mod grpc;

use std::{path::PathBuf, process::exit, str::FromStr};

use clap::Parser;

use log::{error, warn};

pub use grpc::*;

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

    let server = SemanticSearchEngineGrpcServer::new();

    if let Err(e) = server.run().await {
        error!("Could not start the server: {}", e);
        exit(1);
    }

    Ok(())
}
