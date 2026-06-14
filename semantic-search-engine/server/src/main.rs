// Main

use std::{path::PathBuf, str::FromStr};

use clap::Parser;

use log::{info, warn};

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

    info!("Hello, world!");

    Ok(())
}
