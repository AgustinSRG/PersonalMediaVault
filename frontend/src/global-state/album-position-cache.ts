// Cached album position

"use strict";

import { fetchFromLocalStorage, saveIntoLocalStorage, clearLocalStorage } from "@/local-storage/local-storage";

const MAX_CACHE_ALBUM_POS_SIZE = 100;
const LS_KEY_ALBUM_POS_CACHE = "player-album-pos-cache";

interface AlbumPositionCacheEntry {
    id: number;
    pos: number;
}

/**
 * Gets cached current album position
 * @param id Album ID
 * @returns Current cached position
 */
export function getCachedAlbumPosition(id: number): number {
    let cache = fetchFromLocalStorage(LS_KEY_ALBUM_POS_CACHE, [] as AlbumPositionCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    for (const entry of cache) {
        if (!entry || typeof entry !== "object") {
            continue;
        }

        if (entry.id === id) {
            const pos = entry.pos;
            if (typeof pos === "number" && !isNaN(pos) && isFinite(pos) && pos >= 0) {
                return pos;
            } else {
                return 0;
            }
        }
    }

    return 0;
}

/**
 * Sets cached current album position
 * @param id Album ID
 * @param pos Current cached position
 */
export function setCachedAlbumPosition(id: number, pos: number) {
    let cache = fetchFromLocalStorage(LS_KEY_ALBUM_POS_CACHE, [] as AlbumPositionCacheEntry[]);

    if (!cache || !Array.isArray(cache)) {
        cache = [];
    }

    // Remove elements
    cache = cache.filter((e) => {
        if (!e || typeof e !== "object") {
            return false;
        }

        return e.id !== id;
    });

    while (cache.length >= MAX_CACHE_ALBUM_POS_SIZE) {
        cache.shift();
    }

    cache.push({
        id: id,
        pos: pos,
    });

    saveIntoLocalStorage(LS_KEY_ALBUM_POS_CACHE, cache);
}

/**
 * Clears cached initial times
 */
export function clearCachedAlbumPositions() {
    clearLocalStorage(LS_KEY_ALBUM_POS_CACHE);
}
