// Global events manager

"use strict";

import type { Album, MediaData } from "@/api/models";
import { type AlbumListItemMin } from "@/api/models";
import { type SnackBatPosition } from "./pages";
import type { ColorThemeName } from "./app-preferences";
import type { ImageNote } from "@/utils/notes-format";
import type { ImageNodesChangeType } from "./img-notes";
import type { UploadEntryMin } from "./upload";

/**
 * Event triggered when a request results in 401 - Unauthorized
 */
export const EVENT_NAME_UNAUTHORIZED = "unauthorized";

/**
 * Event triggered when a new application version is detected
 */
export const EVENT_NAME_APP_NEW_VERSION = "app-new-version";

/**
 * Event triggered when the auth status starts or stops loading
 */
export const EVENT_NAME_AUTH_LOADING = "auth-status-loading";

/**
 * Event triggered when the authentication status changes
 */
export const EVENT_NAME_AUTH_CHANGED = "auth-status-changed";

/**
 * Event triggered when the authentication status cannot be loaded due to an error
 */
export const EVENT_NAME_AUTH_ERROR = "auth-status-loading-error";

/**
 * Event triggered when the app status changes
 */
export const EVENT_NAME_APP_STATUS_CHANGED = "app-status-update";

/**
 * Event triggered when the locale changes
 */
export const EVENT_NAME_LOCALE_CHANGED = "set-locale";

/**
 * Event triggered when a new locale file is loaded
 */
export const EVENT_NAME_LOADED_LOCALE = "loaded-locale";

/**
 * Event triggered when the theme changes
 */
export const EVENT_NAME_THEME_CHANGED = "theme-changed";

/**
 * Event triggered when the navigation of the current media in the current page changes
 */
export const EVENT_NAME_PAGE_MEDIA_NAV_UPDATE = "page-media-nav-update";

/**
 * Event triggered when a refresh is requested for the random page
 */
export const EVENT_NAME_RANDOM_PAGE_REFRESH = "random-page-refresh";

/**
 * Event triggered when the user requests going to the next media element in the page
 */
export const EVENT_NAME_PAGE_NAV_NEXT = "page-media-nav-next";

/**
 * Event triggered when the user requests going to the previous media element in the page
 */
export const EVENT_NAME_PAGE_NAV_PREV = "page-media-nav-prev";

/**
 * Event triggered when the user request goings to the next media element
 */
export const EVENT_NAME_GO_NEXT = "media-go-next";

/**
 * Event triggered when the user requests going to the previous media element
 */
export const EVENT_NAME_GO_PREV = "media-go-prev";

/**
 * Event triggered when the user requests going to the top of the advanced search page
 */
export const EVENT_NAME_ADVANCED_SEARCH_GO_TOP = "search-go-top";

/**
 * Event triggered when the advanced search container is scrolled
 */
export const EVENT_NAME_ADVANCED_SEARCH_SCROLL = "search-scroll";

/**
 * Event triggered when the user makes changes to the basic metadata of the current media
 * This means albums and pages should reload
 */
export const EVENT_NAME_MEDIA_METADATA_CHANGE = "media-meta-change";

/**
 * Event triggered when the user deletes the current media
 * This means albums and pages should reload
 */
export const EVENT_NAME_MEDIA_DELETE = "media-delete";

/**
 * Event triggered to show a message in the snackbar
 */
export const EVENT_NAME_SNACK_BAR = "snack";

/**
 * Event triggered when the albums list is updated
 */
export const EVENT_NAME_ALBUMS_LIST_UPDATE = "albums-update";

/**
 * Event triggered when the user updates an album, so the list must be re-fetched
 */
export const EVENT_NAME_ALBUMS_CHANGED = "albums-list-change";

/**
 * Event triggered when the loading status for the current album changes
 */
export const EVENT_NAME_CURRENT_ALBUM_LOADING = "current-album-loading";

/**
 * Event triggered when the current album data is updated
 */
export const EVENT_NAME_CURRENT_ALBUM_UPDATED = "current-album-update";

/**
 * Event triggered when the current media position in the current album is updated
 */
export const EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED = "album-pos-update";

/**
 * Event triggered when the next element is pre-fetched
 */
export const EVENT_NAME_NEXT_PRE_FETCH = "album-next-prefetch";

/**
 * Event triggered when the page preferences are updated
 */
export const EVENT_NAME_PAGE_PREFERENCES_UPDATED = "page-preferences-updated";

/**
 * Event emitted when the list of favorite albums changes
 */
export const EVENT_NAME_FAVORITE_ALBUMS_UPDATED = "albums-fav-updated";

/**
 * Event triggered when an album is loaded to put in at the top of the sidebar list
 */
export const EVENT_NAME_ALBUM_SIDEBAR_TOP = "album-sidebar-top";

/**
 * Event triggered when the image notes are updated
 */
export const EVENT_NAME_IMAGE_NOTES_UPDATE = "img-notes-update";

/**
 * Event triggered when the image notes are changed
 */
export const EVENT_NAME_IMAGE_NOTES_CHANGE = "img-notes-change";

/**
 * Event triggered when the image notes are saved
 */
export const EVENT_NAME_IMAGE_NOTES_SAVED = "image-notes-saved";

/**
 * Event triggered when the media loading status changes
 */
export const EVENT_NAME_MEDIA_LOADING = "current-media-loading";

/**
 * Event triggered when the current media data is updated
 */
export const EVENT_NAME_MEDIA_UPDATE = "current-media-update";

/**
 * Event triggered when the current media description is updated
 */
export const EVENT_NAME_MEDIA_DESCRIPTION_UPDATE = "current-media-description-update";

/**
 * Event triggered when a subtitles file is loaded
 */
export const EVENT_NAME_SUBTITLES_UPDATE = "subtitles-update";

/**
 * Event triggered when the subtitles options change
 */
export const EVENT_NAME_SUBTITLES_OPTIONS_CHANGED = "subtitles-options-changed";

/**
 * Event triggered whenever the tags list is loaded or changed
 */
export const EVENT_NAME_TAGS_UPDATE = "tags-update";

/**
 * Event triggered when an upload entry is added
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_NEW = "upload-list-entry-new";

/**
 * Event triggered when an upload entry progresses
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS = "upload-list-entry-progress";

/**
 * Event triggered when an upload entry is ready
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_READY = "upload-list-entry-ready";

/**
 * Event triggered when an upload entry results in an error
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR = "upload-list-entry-error";

/**
 * Event triggered when an upload entry is retried after error
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY = "upload-list-entry-retry";

/**
 * Event triggered when an upload entry is removed
 */
export const EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED = "upload-list-entry-rm";

/**
 * Event triggered when an upload entry list is cleared, meaning it needs full refresh
 */
export const EVENT_NAME_UPLOAD_LIST_CLEAR = "upload-list-clear";

/**
 * Event triggered when the scroll in the home page changed
 */
export const EVENT_NAME_HOME_SCROLL_CHANGED = "home-page-scroll-changed";

/**
 * Mapping of App events
 *
 * Maps each event name to the type of appropriate handler function
 */
export type AppEventsMap = {
    /**
     * Event triggered when a request results in 401 - Unauthorized
     */
    [EVENT_NAME_UNAUTHORIZED]: () => void;

    /**
     * Event triggered when a new application version is detected
     */
    [EVENT_NAME_APP_NEW_VERSION]: () => void;

    /**
     * Event triggered when the auth status starts or stops loading
     * @param loading True if the auth status is being loaded
     */
    [EVENT_NAME_AUTH_LOADING]: (loading: boolean) => void;

    /**
     * Event triggered when the authentication status changes
     * @param locked True if the vault is locked
     * @param username Username
     */
    [EVENT_NAME_AUTH_CHANGED]: (locked: boolean, username: string) => void;

    /**
     * Event triggered when the authentication status cannot be loaded due to an error
     */
    [EVENT_NAME_AUTH_ERROR]: () => void;

    /**
     * Event triggered when the app status changes
     */
    [EVENT_NAME_APP_STATUS_CHANGED]: () => void;

    /**
     * Event triggered when the locale changes
     * @param locale The new locale
     */
    [EVENT_NAME_LOCALE_CHANGED]: (locale: string) => void;

    /**
     * Event triggered when a new locale file is loaded
     * @param locale The loaded locale
     */
    [EVENT_NAME_LOADED_LOCALE]: (locale: string) => void;

    /**
     * Event triggered when the theme changes
     * @param theme The new theme
     */
    [EVENT_NAME_THEME_CHANGED]: (theme: ColorThemeName) => void;

    /**
     * Event triggered when the navigation of the current media in the current page changes
     * @param pageHasPrev True if the page has a previous element
     * @param pageHasNext True if the page has a next element
     */
    [EVENT_NAME_PAGE_MEDIA_NAV_UPDATE]: (pageHasPrev: boolean, pageHasNext: boolean) => void;

    /**
     * Event triggered when a refresh is requested for the random page
     */
    [EVENT_NAME_RANDOM_PAGE_REFRESH]: () => void;

    /**
     * Event triggered when the user requests going to the next media element in the page
     */
    [EVENT_NAME_PAGE_NAV_NEXT]: () => void;

    /**
     * Event triggered when the user requests going to the previous media element in the page
     */
    [EVENT_NAME_PAGE_NAV_PREV]: () => void;

    /**
     * Event triggered when the user request goings to the next media element
     */
    [EVENT_NAME_GO_NEXT]: () => void;

    /**
     * Event triggered when the user requests going to the previous media element
     */
    [EVENT_NAME_GO_PREV]: () => void;

    /**
     * Event triggered when the user requests going to the top of the advanced search page
     */
    [EVENT_NAME_ADVANCED_SEARCH_GO_TOP]: () => void;

    /**
     * Event triggered when the advanced search container is scrolled
     * @param e The scroll event
     */
    [EVENT_NAME_ADVANCED_SEARCH_SCROLL]: (e: Event) => void;

    /**
     * Event triggered when the user makes changes to the basic metadata of the current media
     * This means albums and pages should reload
     */
    [EVENT_NAME_MEDIA_METADATA_CHANGE]: () => void;

    /**
     * Event triggered when the user deletes the current media
     * This means albums and pages should reload
     */
    [EVENT_NAME_MEDIA_DELETE]: () => void;

    /**
     * Event triggered to show a message in the snackbar
     * @param message The message
     * @param position The position
     */
    [EVENT_NAME_SNACK_BAR]: (message: string, position?: SnackBatPosition) => void;

    /**
     * Event triggered when the albums list is updated
     * @param albumsMap The updated albums map
     */
    [EVENT_NAME_ALBUMS_LIST_UPDATE]: (albumsMap: Map<number, AlbumListItemMin>) => void;

    /**
     * Event triggered when the user updates an album, so the list must be re-fetched
     */
    [EVENT_NAME_ALBUMS_CHANGED]: () => void;

    /**
     * Event triggered when the loading status for the current album changes
     * @param loading True if the album is loading
     */
    [EVENT_NAME_CURRENT_ALBUM_LOADING]: (loading: boolean) => void;

    /**
     * Event triggered when the current album data is updated
     * @param album The album data
     */
    [EVENT_NAME_CURRENT_ALBUM_UPDATED]: (album: Album | null) => void;

    /**
     * Event triggered when the current media position in the current album is updated
     */
    [EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED]: () => void;

    /**
     * Event triggered when the next element is pre-fetched
     */
    [EVENT_NAME_NEXT_PRE_FETCH]: () => void;

    /**
     * Event triggered when the page preferences are updated
     */
    [EVENT_NAME_PAGE_PREFERENCES_UPDATED]: () => void;

    /**
     * Event emitted when the list of favorite albums changes
     */
    [EVENT_NAME_FAVORITE_ALBUMS_UPDATED]: () => void;

    /**
     * Event triggered when an album is loaded to put in at the top of the sidebar list
     * @param albumId The album ID
     */
    [EVENT_NAME_ALBUM_SIDEBAR_TOP]: (albumId: number) => void;

    /**
     * Event triggered when the current media data is updated
     */
    [EVENT_NAME_IMAGE_NOTES_UPDATE]: () => void;

    /**
     * Event triggered when the image notes are changed
     * @param mode The change type
     * @param note The image note (null when removing)
     * @param index The image note index
     */
    [EVENT_NAME_IMAGE_NOTES_CHANGE]: (type: ImageNodesChangeType, note: ImageNote | null, index?: number) => void;

    /**
     * Event triggered when the image notes are saved
     */
    [EVENT_NAME_IMAGE_NOTES_SAVED]: () => void;

    /**
     * Event triggered when the media loading status changes
     * @param loading True if the media is loading
     */
    [EVENT_NAME_MEDIA_LOADING]: (loading: boolean) => void;

    /**
     * Event triggered when the current media data is updated
     * @param media The media data
     */
    [EVENT_NAME_MEDIA_UPDATE]: (media: MediaData | null) => void;

    /**
     * Event triggered when the current media description is updated
     * @param source The source of the event (editor = From player editor, widget = From description widget)
     */
    [EVENT_NAME_MEDIA_DESCRIPTION_UPDATE]: (source: "editor" | "widget") => void;

    /**
     * Event triggered when a subtitles file is loaded
     */
    [EVENT_NAME_SUBTITLES_UPDATE]: () => void;

    /**
     * Event triggered when the subtitles options change
     */
    [EVENT_NAME_SUBTITLES_OPTIONS_CHANGED]: () => void;

    /**
     * Event triggered whenever the tags list is loaded or changed
     * @param version The version of the tags
     */
    [EVENT_NAME_TAGS_UPDATE]: (version: number) => void;

    /**
     * Event triggered when an upload entry is added
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_NEW]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry progresses
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry is ready
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_READY]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry results in an error
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry is retried after error
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry is removed
     * @param entry The upload entry
     */
    [EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED]: (entry: UploadEntryMin) => void;

    /**
     * Event triggered when an upload entry list is cleared, meaning it needs full refresh
     */
    [EVENT_NAME_UPLOAD_LIST_CLEAR]: () => void;

    /**
     * Event triggered when the scroll in the home page changed
     */
    [EVENT_NAME_HOME_SCROLL_CHANGED]: () => void;
};

/**
 * Map of event handlers
 */
const AppEventsHandlers: Map<keyof AppEventsMap, AppEventsMap[keyof AppEventsMap][]> = new Map();

/**
 * Adds a listener for a global application event
 * @param eventName The event name
 * @param listener The event listener
 */
export function addAppEventListener<K extends keyof AppEventsMap>(eventName: K, listener: AppEventsMap[K]) {
    const handlers = AppEventsHandlers.get(eventName) as AppEventsMap[K][];

    if (handlers) {
        handlers.push(listener);
    } else {
        AppEventsHandlers.set(eventName, [listener]);
    }
}

/**
 * Removes a listener for a global application event
 * @param eventName The event name
 * @param listener The event listener
 */
export function removeAppEventListener<K extends keyof AppEventsMap>(eventName: K, listener: AppEventsMap[K]) {
    const handlers = AppEventsHandlers.get(eventName) as AppEventsMap[K][];

    if (handlers) {
        const i = handlers.indexOf(listener);

        if (i >= 0) {
            handlers.splice(i, 1);

            if (handlers.length === 0) {
                AppEventsHandlers.delete(eventName);
            }
        }
    }
}

/**
 * Emits a global application event
 * @param eventName The event name
 * @param args The event arguments
 */
export function emitAppEvent<K extends keyof AppEventsMap>(eventName: K, ...args: Parameters<AppEventsMap[K]>) {
    const handlers = AppEventsHandlers.get(eventName) as AppEventsMap[K][];

    if (handlers) {
        for (const handler of handlers) {
            try {
                handler.call(null, ...args);
            } catch (ex) {
                console.error(ex);
            }
        }
    }
}
