// CLIP model loader

use crate::ClipModelLoadError;
use open_clip_inference::Clip;
use ort::ep::{CUDA, ROCm, TensorRT};
use std::path::Path;

/// Loaded CLIP model
pub struct LoadedClipModel {
    clip: Clip,
}

impl LoadedClipModel {
    /// Loads the model from a directory
    pub fn load(path: &Path) -> Result<LoadedClipModel, ClipModelLoadError> {
        let clip = Clip::from_local_dir(path)
            .with_execution_providers(&[
                TensorRT::default().build(),
                CUDA::default().build(),
                ROCm::default().build(),
            ])
            .build()?;

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
