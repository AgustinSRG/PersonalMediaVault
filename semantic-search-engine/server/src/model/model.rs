// CLIP model loader

use std::{path::PathBuf, sync::Arc};

use open_clip_inference::Clip;

use crate::{ClipModelLoadError, api::ClipModelMetadataResponse};

/// Loaded CLIP model
pub struct LoadedClipModel {
    clip: Arc<Clip>,
}

impl LoadedClipModel {
    /// Loads the model from a directory
    pub fn load(path: &PathBuf) -> Result<LoadedClipModel, ClipModelLoadError> {
        let clip = Clip::from_local_dir(path).build()?;

        Ok(LoadedClipModel {
            clip: Arc::new(clip),
        })
    }

    /// Gets the CLIP instance
    pub fn get_clip(&self) -> Arc<Clip> {
        self.clip.clone()
    }

    // Gets dimensions of embedding vectors
    pub fn get_embed_dim(&self) -> u32 {
        self.clip.vision.config.model_cfg.embed_dim as u32
    }
}
