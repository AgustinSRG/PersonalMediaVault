// App preferences

"use strict";

import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "./albums";
import { AppEvents } from "./app-events";
import { clearLocalStorage, fetchFromLocalStorage, fetchFromLocalStorageCache, saveIntoLocalStorage } from "../utils/local-storage";
import { AlbumListItemMin } from "@/api/models";

export type ColorThemeName = "light" | "dark";

const LS_KEY_THEME = "app-pref-theme";

/**
 * Event triggered when the theme changes
 */
export const EVENT_NAME_THEME_CHANGED = "theme-changed";

/**
 * Gets color theme
 * @returns The theme name
 */
export function getTheme(): ColorThemeName {
    return fetchFromLocalStorageCache(LS_KEY_THEME, "dark");
}

/**
 * Sets color theme
 * @param theme The theme name
 */
export function setTheme(theme: ColorThemeName) {
    saveIntoLocalStorage(LS_KEY_THEME, theme);
    AppEvents.Emit(EVENT_NAME_THEME_CHANGED, theme);
}

const LS_KEY_FAVORITE_ALBUMS = "app-pref-albums-fav";

export const EVENT_NAME_FAVORITE_ALBUMS_UPDATED = "albums-fav-updated";

/**
 * Checks if album is favorite
 * @param id Album ID
 * @returns True if the album is in favorites
 */
export function albumIsFavorite(id: number) {
    const idString = id + "";
    let favorites = fetchFromLocalStorage(LS_KEY_FAVORITE_ALBUMS, []);

    if (!favorites || !Array.isArray(favorites)) {
        favorites = [];
    }

    return favorites.includes(idString);
}

/**
 * Gets the list of favorites albums
 * @returns List of favorites albums
 */
export function getAlbumFavoriteList() {
    let favorites = fetchFromLocalStorage(LS_KEY_FAVORITE_ALBUMS, []);

    if (!favorites || !Array.isArray(favorites)) {
        favorites = [];
    }

    return favorites;
}

/**
 * Adds album to favorites
 * @param id Album ID
 */
export function albumAddFav(id: number) {
    const idString = id + "";
    let favorites = fetchFromLocalStorage(LS_KEY_FAVORITE_ALBUMS, []);

    if (!favorites || !Array.isArray(favorites)) {
        favorites = [];
    }

    if (!favorites.includes(idString)) {
        favorites.push(idString);
        saveIntoLocalStorage(LS_KEY_FAVORITE_ALBUMS, favorites);
        AppEvents.Emit(EVENT_NAME_FAVORITE_ALBUMS_UPDATED);
    }
}

/**
 * Removes albums from favorites
 * @param id Album ID
 */
export function albumRemoveFav(id: number) {
    const idString = id + "";
    let favorites = fetchFromLocalStorage(LS_KEY_FAVORITE_ALBUMS, []);

    if (!favorites || !Array.isArray(favorites)) {
        favorites = [];
    }

    const index = favorites.indexOf(idString);
    if (index >= 0) {
        favorites.splice(index, 1);
        saveIntoLocalStorage(LS_KEY_FAVORITE_ALBUMS, favorites);
        AppEvents.Emit(EVENT_NAME_FAVORITE_ALBUMS_UPDATED);
    }
}

AppEvents.AddEventListener(EVENT_NAME_ALBUMS_LIST_UPDATE, (albums: Map<number, AlbumListItemMin>) => {
    // Remove favorite albums removed from the vault
    let favorites = fetchFromLocalStorage(LS_KEY_FAVORITE_ALBUMS, []);

    if (!favorites || !Array.isArray(favorites)) {
        favorites = [];
    } else {
        favorites = favorites.filter((id) => {
            return albums.has(parseInt(id, 10));
        });
        saveIntoLocalStorage(LS_KEY_FAVORITE_ALBUMS, favorites);
    }
});

/**
 * Event triggered when an album is loaded to put in at the top of the sidebar list
 */
export const EVENT_NAME_ALBUM_SIDEBAR_TOP = "album-sidebar-top";

const LS_KEY_ALBUMS_ORDER = "app-pref-albums-order";

const ALBUM_ORDER_MAP_SIZE_LIMIT = 100;

/**
 * Gets albums ordering map.
 * Assigns each album the timestamp it was last loaded in.
 * @returns
 */
export function getAlbumsOrderMap(): { [id: string]: number } {
    let m = fetchFromLocalStorage(LS_KEY_ALBUMS_ORDER, Object.create(null));

    if (!m || typeof m !== "object") {
        m = Object.create(null);
    }

    return m;
}

AppEvents.AddEventListener(EVENT_NAME_CURRENT_ALBUM_UPDATED, () => {
    if (!AlbumsController.CurrentAlbumData) {
        return;
    }

    const idString = AlbumsController.CurrentAlbumData.id + "";
    const m = getAlbumsOrderMap();

    m[idString] = Date.now();

    while (Object.keys(m).length > ALBUM_ORDER_MAP_SIZE_LIMIT) {
        let oldest: string;
        let oldestVal = Date.now();

        for (const albumIdString of Object.keys(m)) {
            const v = m[albumIdString];

            if (v < oldestVal) {
                oldestVal = v;
                oldest = albumIdString;
            }
        }

        delete m[oldest];
    }

    saveIntoLocalStorage(LS_KEY_ALBUMS_ORDER, m);

    AppEvents.Emit(EVENT_NAME_ALBUM_SIDEBAR_TOP, AlbumsController.CurrentAlbumData.id);
});

const LS_KEY_PAGE_SETTINGS = "app-pref-page-settings";

/**
 * Page preferences
 */
export interface PagePreferences {
    /**
     * Max items per page
     */
    pageSize: number;

    /**
     * Row size when expanded
     */
    rowSize: number;

    /**
     * Row size when split mode
     */
    rowSizeMin: number;

    /**
     * Min size of items
     */
    minItemSize: number;

    /**
     * Max size of items
     */
    maxItemSize: number;

    /**
     * Display titles?
     */
    displayTitles: boolean;

    /**
     * Rounded corners?
     */
    roundedCorners: boolean;

    /**
     * Padding for items
     */
    padding: number;
}

/**
 * Event triggered when the page preferences are updated
 */
export const EVENT_NAME_PAGE_PREFERENCES_UPDATED = "page-preferences-updated";

/**
 * Default page preferences
 */
const DEFAULT_PAGE_PREFERENCES: PagePreferences = {
    pageSize: 25,
    rowSize: 5,
    rowSizeMin: 2,
    minItemSize: 64,
    maxItemSize: 464,
    displayTitles: true,
    roundedCorners: true,
    padding: 18,
};

/**
 * Gets the page preferences
 * @param page The page name
 * @returns The page preferences
 */
export function getPagePreferences(page: string): PagePreferences {
    if (window.innerWidth > 1000) {
        DEFAULT_PAGE_PREFERENCES.rowSize = Math.max(2, Math.round((window.innerWidth - 240) / 250));
    } else {
        DEFAULT_PAGE_PREFERENCES.rowSize = Math.max(2, Math.round(window.innerWidth / 250));
    }

    DEFAULT_PAGE_PREFERENCES.pageSize = Math.min(256, DEFAULT_PAGE_PREFERENCES.rowSize * 6);

    let preferences =
        fetchFromLocalStorageCache(
            LS_KEY_PAGE_SETTINGS + "-" + page,
            fetchFromLocalStorageCache(LS_KEY_PAGE_SETTINGS, DEFAULT_PAGE_PREFERENCES),
        ) || DEFAULT_PAGE_PREFERENCES;

    if (typeof preferences !== "object" || Array.isArray(preferences)) {
        preferences = DEFAULT_PAGE_PREFERENCES;
    }

    for (const key of Object.keys(DEFAULT_PAGE_PREFERENCES)) {
        if (preferences[key] === undefined || typeof preferences[key] !== typeof DEFAULT_PAGE_PREFERENCES[key]) {
            preferences[key] = DEFAULT_PAGE_PREFERENCES[key];
        }
    }

    if (isNaN(preferences.pageSize) || preferences.pageSize < 1 || preferences.pageSize > 256) {
        preferences.pageSize = DEFAULT_PAGE_PREFERENCES.pageSize;
    }

    if (isNaN(preferences.rowSize) || preferences.rowSize < 1 || preferences.rowSize > 256) {
        preferences.rowSize = DEFAULT_PAGE_PREFERENCES.rowSize;
    }

    if (isNaN(preferences.rowSizeMin) || preferences.rowSizeMin < 1 || preferences.rowSizeMin > 256) {
        preferences.rowSizeMin = DEFAULT_PAGE_PREFERENCES.rowSizeMin;
    }

    if (isNaN(preferences.minItemSize) || preferences.minItemSize < 1 || preferences.minItemSize > 1000) {
        preferences.minItemSize = DEFAULT_PAGE_PREFERENCES.minItemSize;
    }

    if (isNaN(preferences.maxItemSize) || preferences.maxItemSize < 1 || preferences.maxItemSize > 1000) {
        preferences.maxItemSize = DEFAULT_PAGE_PREFERENCES.maxItemSize;
    }

    if (isNaN(preferences.padding) || preferences.padding < 0 || preferences.padding > 32) {
        preferences.maxItemSize = DEFAULT_PAGE_PREFERENCES.maxItemSize;
    }

    return preferences;
}

/**
 * Sets the page preferences
 * @param page The page name
 * @param preferences The page preferences
 */
export function setPagePreferences(page: string, preferences: PagePreferences) {
    saveIntoLocalStorage(LS_KEY_PAGE_SETTINGS + "-" + page, preferences);
    AppEvents.Emit(EVENT_NAME_PAGE_PREFERENCES_UPDATED);
}

/**
 * Resets page preferences
 * @param page Page name
 */
export function resetPagePreferences(page: string) {
    clearLocalStorage(LS_KEY_PAGE_SETTINGS + page);
    clearLocalStorage(LS_KEY_PAGE_SETTINGS);
    AppEvents.Emit(EVENT_NAME_PAGE_PREFERENCES_UPDATED);
}

const LS_KEY_LAST_USED_TAGS = "app-last-used-tags";

const LAST_USED_TAGS_LIMIT = 10;

/**
 * Gets the list of last used tags
 */
export function getLastUsedTags(): number[] {
    let r = fetchFromLocalStorageCache(LS_KEY_LAST_USED_TAGS, []);

    if (!Array.isArray(r)) {
        r = [];
    }

    return r;
}

/**
 * Updates the last used tags
 * @param tag The used tag
 */
export function setLastUsedTag(tag: number) {
    let r = fetchFromLocalStorage(LS_KEY_LAST_USED_TAGS, []);

    if (!Array.isArray(r)) {
        r = [];
    }

    if (r.includes(tag)) {
        r = r.filter((t) => t !== tag);
    }

    r.unshift(tag);

    while (r.length > LAST_USED_TAGS_LIMIT) {
        r.pop();
    }

    saveIntoLocalStorage(LS_KEY_LAST_USED_TAGS, r);
}
