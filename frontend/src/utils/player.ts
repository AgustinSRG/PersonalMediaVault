// Player utils

"use strict";

// Player load status
//  - loading: Loading the media
//  - 200: OK (loaded)
//  - none: No media selected
/// - 404: Media not found
export type PlayerLoadStatus = "loading" | "200" | "none" | "404";

// Player play feedback
// For audio and video players
export type PlayerPlayFeedbackType = "" | "play" | "pause";
