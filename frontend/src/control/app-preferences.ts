// App preferences

"use strict";

import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE, EVENT_NAME_CURRENT_ALBUM_UPDATED } from "./albums";
import { AppEvents } from "./app-events";
import { fetchFromLocalStorage, fetchFromLocalStorageCache, saveIntoLocalStorage } from "../utils/local-storage";
import { AlbumListItemMin } from "@/api/models";

export type ColorThemeName = "light" | "dark";

/**
 * Gets default theme, by checking the browser settings
 * @returns The theme name
 */
function defaultBrowserTheme(): ColorThemeName {
    if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
        return "dark";
    } else if (window.matchMedia) {
        return "light";
    } else {
        return "dark";
    }
}

const LS_KEY_THEME = "app-pref-theme";

const Theme = {
    loaded: false,
    value: "light" as ColorThemeName,
};

/**
 * Event triggered when the theme changes
 */
export const EVENT_NAME_THEME_CHANGED = "theme-changed";

/**
 * Gets color theme
 * @returns The theme name
 */
export function getTheme(): ColorThemeName {
    if (!Theme.loaded) {
        Theme.value = fetchFromLocalStorage(LS_KEY_THEME, defaultBrowserTheme());
        Theme.loaded = true;
    }

    return Theme.value;
}

/**
 * Sets color theme
 * @param theme The theme name
 */
export function setTheme(theme: ColorThemeName) {
    Theme.value = theme;
    saveIntoLocalStorage(LS_KEY_THEME, theme);
    AppEvents.Emit(EVENT_NAME_THEME_CHANGED, theme);
}

const LS_KEY_LANGUAGE = "app-pref-lang";

const Language = {
    loaded: false,
    value: "en",
};

/**
 * Event triggered when the locale changes
 */
export const EVENT_NAME_LOCALE_CHANGED = "set-locale";

/**
 * Gets the language
 * @returns Language
 */
export function getLanguage(): string {
    if (!Language.loaded) {
        Language.value = fetchFromLocalStorage(LS_KEY_LANGUAGE, navigator.language || "en");
        Language.loaded = true;
    }

    return Language.value;
}

/**
 * Sets the language
 * @param lang Language
 */
export function setLanguage(lang: string) {
    Language.value = lang;
    saveIntoLocalStorage(LS_KEY_LANGUAGE, lang);
    AppEvents.Emit(EVENT_NAME_LOCALE_CHANGED, lang);
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

const LS_KEY_PAGE_MAX_ITEMS = "app-pref-page-max-items";

/**
 * Event triggered when the max items per page setting changes
 */
export const EVENT_NAME_PAGE_SIZE_UPDATED = "page-size-pref-updated";

/**
 * Gets max items per page setting
 * @returns Max items per page
 */
export function getPageMaxItems(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_PAGE_MAX_ITEMS, 25)) || 0;
}

/**
 * Sets max items per page setting
 * @returns Max items per page
 */
export function setPageMaxItems(maxItems: number) {
    saveIntoLocalStorage(LS_KEY_PAGE_MAX_ITEMS, maxItems);
    AppEvents.Emit(EVENT_NAME_PAGE_SIZE_UPDATED);
}

/**
 * Event triggered when the page items settings change
 */
export const EVENT_NAME_PAGE_ITEMS_UPDATED = "page-items-pref-updated";

const LS_KEY_PAGE_ITEMS_SIZE = "app-pref-page-items-size";

/**
 * Gets page items size class
 * @returns The size class
 */
export function getPageItemsSize(): string {
    return fetchFromLocalStorageCache(LS_KEY_PAGE_ITEMS_SIZE, "normal") + "";
}

/**
 * Sets page items size class
 * @param size The size class
 */
export function setPageItemsSize(size: string) {
    saveIntoLocalStorage(LS_KEY_PAGE_ITEMS_SIZE, size);
    AppEvents.Emit(EVENT_NAME_PAGE_ITEMS_UPDATED);
}

const LS_KEY_PAGE_ITEMS_FIT = "app-pref-page-items-fit";

/**
 * Gets number of items to fit in a row of a page
 * @returns Number of items to fit
 */
export function getPageItemsFit(): number {
    return Number(fetchFromLocalStorageCache(LS_KEY_PAGE_ITEMS_FIT, 5)) || 0;
}

/**
 * Sets number of items to fit in a row of a page
 * @param fit Number of items to fit
 */
export function setPageItemsFit(fit: number) {
    saveIntoLocalStorage(LS_KEY_PAGE_ITEMS_FIT, fit);
    AppEvents.Emit(EVENT_NAME_PAGE_ITEMS_UPDATED);
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
        r = r.filter(t => t !== tag);
    }

    r.unshift(tag);

    while (r.length > LAST_USED_TAGS_LIMIT) {
        r.pop();
    }

    saveIntoLocalStorage(LS_KEY_LAST_USED_TAGS, r);
}
