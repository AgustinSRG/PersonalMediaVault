// Global state initialization.
// This file is responsible of initializing global
// loaders to ensure they are ready as fast as possible

"use strict";

import { initializeAuthentication } from "./global-state/auth";
import { initializeNavigation } from "./global-state/navigation";
import { initializeAlbums } from "./global-state/albums";
import { initializeTags } from "./global-state/tags";
import { initializeAlbum } from "./global-state/album";
import { initializeMedia } from "./global-state/media";

/**
 * Initializes the global state that may need initialization
 */
export function initializeGlobalState() {
    // Initialize authentication
    initializeAuthentication();

    // Initialize navigation status
    initializeNavigation();

    // Initialize loader for global albums list
    initializeAlbums();

    // Initialize loader for global tags list
    initializeTags();

    // Initializes loader and state for the current album
    initializeAlbum();

    // Initializes loader and state for the current media
    initializeMedia();
}
