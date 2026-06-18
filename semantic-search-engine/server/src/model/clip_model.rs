// CLIP model loader

use std::path::Path;

use open_clip_inference::Clip;

use crate::ClipModelLoadError;

/// Loaded CLIP model
pub struct LoadedClipModel {
    clip: Clip,
}

impl LoadedClipModel {
    /// Loads the model from a directory
    pub fn load(path: &Path) -> Result<LoadedClipModel, ClipModelLoadError> {
        let clip = Clip::from_local_dir(path).build()?;

        Ok(LoadedClipModel { clip })
    }

    /// Gets the CLIP instance
    pub fn get_clip(&self) -> &Clip {
        &self.clip
    }

    // Gets dimensions of embedding vectors
    pub fn get_embed_dim(&self) -> u32 {
        self.clip.vision.config.model_cfg.embed_dim as u32
    }
}
