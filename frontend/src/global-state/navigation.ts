// Navigation global status

"use strict";

import { generateURIQuery } from "@/utils/api";
import { getParameterByName } from "@/utils/query-parameter";
import { isAlbumsListLoading } from "./albums";
import { getCurrentAlbumData, getCurrentAlbumId } from "./album";
import { emitAppEvent, EVENT_NAME_NAV_STATUS_CHANGED, EVENT_NAME_MEDIA_DELETE } from "./app-events";
import { tryPreventableExit } from "./exit-prevent";
import { getCachedAlbumPosition } from "./album-position-cache";

/**
 * Layout mode
 */
export type NavigationStatusLayout = "initial" | "album" | "media" | "media-split";

/**
 * Focus target
 */
export type NavigationStatusFocus = "right" | "left";

/**
 * Page
 */
export type NavigationStatusPage = "home" | "media" | "random" | "albums" | "upload" | "search";

/**
 * Page list
 */
const PAGE_LIST: NavigationStatusPage[] = ["home", "media", "random", "albums", "upload", "search"];

/**
 * Navigation status
 */
export type NavigationStatus = {
    /**
     * Current layout
     */
    layout: NavigationStatusLayout;

    /**
     * Current focus
     */
    focus: NavigationStatusFocus;

    /**
     * Current page
     */
    page: NavigationStatusPage;

    /**
     * Current search query
     */
    search: string;

    /**
     * Current search params
     */
    searchParams: string;

    /**
     * Random seed
     */
    randomSeed: number;

    /**
     * True for split mode, false for single view mode
     */
    listSplitMode: boolean;

    /**
     * Current media
     */
    media: number;

    /**
     * Current album
     */
    album: number;

    /**
     * Current home page group
     */
    homePageGroup: number;
};

/**
 * Navigation state
 */
const NavigationState: NavigationStatus = {
    /**
     * Current layout
     */
    layout: "initial" as NavigationStatusLayout,

    /**
     * Current focus
     */
    focus: "left" as NavigationStatusFocus,

    /**
     * Current page
     */
    page: "home" as NavigationStatusPage,

    /**
     * Current search query
     */
    search: "",

    /**
     * Current search params
     */
    searchParams: "",

    /**
     * Random seed
     */
    randomSeed: 0,

    /**
     * True for split mode, false for single view mode
     */
    listSplitMode: true,

    /**
     * Current media
     */
    media: -1,

    /**
     * Current album
     */
    album: -1,

    /**
     * Current home page group
     */
    homePageGroup: -1,
};

/**
 * Gets navigation state
 * @returns The navigation state
 */
export function getNavigationStatus(): Readonly<NavigationStatus> {
    return NavigationState;
}

/**
 * Checks if player is visible
 * @returns True if player is visible
 */
export function isPlayerVisible(): boolean {
    switch (NavigationState.layout) {
        case "album":
        case "media-split":
        case "media":
            return true;
        default:
            return false;
    }
}

/**
 * Checks if page is visible
 * @returns True if page is visible
 */
export function isPageVisible(): boolean {
    switch (NavigationState.layout) {
        case "initial":
        case "media-split":
            return true;
        default:
            return false;
    }
}

/**
 * Loads the URL params,
 * updating the navigation status from them
 */
function loadUrlParams() {
    const media = getParameterByName("media");
    if (media) {
        const mediaId = parseInt(media);

        if (!isNaN(mediaId) && mediaId >= 0) {
            NavigationState.media = mediaId;
        } else {
            NavigationState.media = -1;
        }
    } else {
        NavigationState.media = -1;
    }

    const album = getParameterByName("album");
    if (album) {
        const albumId = parseInt(album);

        if (!isNaN(albumId) && albumId >= 0) {
            NavigationState.album = albumId;
        } else {
            NavigationState.album = -1;
        }
    } else {
        NavigationState.album = -1;
    }

    const page = getParameterByName("page") as NavigationStatusPage;

    if (page && PAGE_LIST.includes(page)) {
        NavigationState.page = page;
    } else {
        NavigationState.page = "home";
    }

    const search = getParameterByName("search");

    if (search) {
        NavigationState.search = search;
    } else {
        NavigationState.search = "";
    }

    NavigationState.homePageGroup = -1;

    if (NavigationState.page === "random") {
        NavigationState.randomSeed = Math.floor(parseInt(getParameterByName("seed") || "0", 10)) || Date.now();
    } else if (NavigationState.page === "home" && NavigationState.media >= 0) {
        const group = getParameterByName("g");
        if (group) {
            const groupId = parseInt(group);

            if (!isNaN(groupId) && groupId >= 0) {
                NavigationState.homePageGroup = groupId;
            }
        }
    }

    const searchParams = getParameterByName("sp");

    if (searchParams) {
        NavigationState.searchParams = searchParams;
    } else {
        NavigationState.searchParams = "";
    }

    const split = getParameterByName("split");

    NavigationState.listSplitMode = split === "yes";
}

/**
 * Updates layout mode based on current status variables,
 * like: page, media, album, split mode
 */
function updateLayout() {
    if (NavigationState.media >= 0) {
        if (NavigationState.album >= 0) {
            // Media with album list
            NavigationState.layout = "album";
        } else if (NavigationState.listSplitMode) {
            // Media with list
            NavigationState.layout = "media-split";
        } else {
            // Media alone
            NavigationState.layout = "media";
        }
    } else if (NavigationState.album >= 0) {
        NavigationState.layout = "album";
    } else {
        // Just initial, sidebar + List
        NavigationState.layout = "initial";
    }
}

/**
 * Generates an URL from the current status
 * @returns The current URL
 */
function getCurrentUrl(): string {
    const params: any = Object.create(null);

    if (NavigationState.media >= 0) {
        params["media"] = NavigationState.media + "";
    }

    if (NavigationState.album >= 0) {
        params["album"] = NavigationState.album + "";
    }

    if (NavigationState.page && NavigationState.page !== "home") {
        params["page"] = NavigationState.page;
    }

    if (NavigationState.page === "home") {
        if (NavigationState.homePageGroup >= 0 && NavigationState.media >= 0) {
            params["g"] = NavigationState.homePageGroup + "";
        }
    }

    if (NavigationState.search) {
        params["search"] = NavigationState.search;
    }

    if (NavigationState.searchParams) {
        params["sp"] = NavigationState.searchParams;
    }

    if (NavigationState.randomSeed && NavigationState.page === "random") {
        params["seed"] = NavigationState.randomSeed + "";
    }

    if (NavigationState.listSplitMode) {
        params["split"] = "yes";
    }

    return window.location.protocol + "//" + window.location.host + window.location.pathname + generateURIQuery(params);
}

/**
 * Updates current URL
 * @param replaceState True if the new URL must replace the old URL
 */
export function navigationUpdateURL(replaceState?: boolean) {
    if (NavigationState.album >= 0 && NavigationState.media < 0) {
        if (isAlbumsListLoading()) {
            return;
        }

        if (getCurrentAlbumId() !== NavigationState.album) {
            return;
        }

        const currentAlbumData = getCurrentAlbumData();

        if (currentAlbumData && currentAlbumData.list.length > 0) {
            return;
        }
    }

    if (history.pushState) {
        const newURL = getCurrentUrl();
        // Update URL
        if (newURL !== location.href) {
            if (replaceState) {
                window.history.replaceState({ path: newURL }, "", newURL);
            } else {
                window.history.pushState({ path: newURL }, "", newURL);
            }
        }
    }
}

/**
 * Call when status updates
 * Updates layout, emits event and updates URL
 * @param replaceState True to replace the old URL with the new one
 */
function onStatusUpdate(replaceState?: boolean) {
    updateLayout();

    emitAppEvent(EVENT_NAME_NAV_STATUS_CHANGED, NavigationState);

    navigationUpdateURL(replaceState);
}

/**
 * Navigates to a page
 * @param page The page to navigate to
 */
export function navigationGoToPage(page: NavigationStatusPage) {
    tryPreventableExit(() => {
        NavigationState.page = page;

        if (NavigationState.page === "random") {
            NavigationState.randomSeed = Date.now();
        }

        NavigationState.album = -1;

        NavigationState.homePageGroup = -1;

        if (NavigationState.media >= 0) {
            NavigationState.listSplitMode = true;
        }

        NavigationState.searchParams = "";

        updateLayout();

        NavigationState.focus = "right";

        onStatusUpdate();
    });
}

/**
 * Navigates to a page
 * Sets split mode only if there is a media selected, and no split enabled
 * @param page The page to navigate to
 * @param searchParams Search parameters
 */
export function navigationGoToPageConditionalSplit(page: NavigationStatusPage, searchParams?: string) {
    tryPreventableExit(() => {
        const changedPage = NavigationState.page !== page;

        NavigationState.page = page;

        if (NavigationState.page === "random") {
            NavigationState.randomSeed = Date.now();
        }

        if (NavigationState.media !== -1 && NavigationState.album === -1 && (!NavigationState.listSplitMode || changedPage)) {
            NavigationState.listSplitMode = true;
        } else {
            NavigationState.media = -1;
            NavigationState.listSplitMode = false;
        }

        NavigationState.album = -1;

        NavigationState.searchParams = searchParams || "";

        updateLayout();

        NavigationState.focus = "right";

        onStatusUpdate();
    });
}

/**
 * Expands the page, closing the player
 */
export function navigationExpandPage() {
    tryPreventableExit(() => {
        NavigationState.album = -1;
        NavigationState.media = -1;
        NavigationState.homePageGroup = -1;
        NavigationState.listSplitMode = false;

        updateLayout();

        NavigationState.focus = "right";

        onStatusUpdate();
    });
}

/**
 * Call when the current media is deleted
 */
export function navigationOnDeleteMedia() {
    NavigationState.media = -1;
    NavigationState.homePageGroup = -1;

    updateLayout();

    NavigationState.focus = "right";

    emitAppEvent(EVENT_NAME_MEDIA_DELETE);

    onStatusUpdate();
}

/**
 * Generates a random new seed
 */
export function navigationRefreshSeed() {
    NavigationState.randomSeed = Date.now();
    onStatusUpdate();
}

/**
 * Changes the search query
 * @param search The search query
 * @param forced True to force the page change to the search results
 */
export function navigationGoToSearch(search: string, forced?: boolean) {
    NavigationState.search = search;

    if (NavigationState.page === "random") {
        NavigationState.randomSeed = Date.now();
    }

    if (NavigationState.search) {
        if (forced || NavigationState.page !== "random") {
            NavigationState.page = "media";
        }

        if (NavigationState.media >= 0) {
            NavigationState.listSplitMode = true;
        }

        NavigationState.album = -1;

        NavigationState.searchParams = "";

        updateLayout();

        NavigationState.focus = "right";
    }

    onStatusUpdate();
}

/**
 * Clears current search query
 */
export function navigationClearSearch() {
    NavigationState.search = "";

    updateLayout();

    NavigationState.focus = "right";

    onStatusUpdate();
}

/**
 * Navigates to a media asset
 * @param mediaId The media ID
 * @param split True to use split mode
 * @param group ID of the home group
 */
export function navigationClickOnMedia(mediaId: number, split: boolean, group?: number) {
    tryPreventableExit(() => {
        NavigationState.media = mediaId;

        if (split) {
            NavigationState.listSplitMode = true;

            if (typeof group === "number" && group >= 0) {
                NavigationState.homePageGroup = group;
            }
        }

        updateLayout();

        NavigationState.focus = "left";

        onStatusUpdate();
    });
}

/**
 * Navigates to an album
 * @param albumId The album ID
 */
export function navigationClickOnAlbum(albumId: number) {
    tryPreventableExit(() => {
        NavigationState.album = albumId;
        NavigationState.media = -1;
        NavigationState.homePageGroup = -1;
        NavigationState.page = "home";

        NavigationState.listSplitMode = false;

        updateLayout();

        NavigationState.focus = "left";

        onStatusUpdate();
    });
}

/**
 * Navigates to an album, while keeping the current media
 * @param albumId The album ID
 * @param mediaId The media ID
 */
export function navigationClickOnAlbumByMedia(albumId: number, mediaId: number) {
    NavigationState.album = albumId;
    NavigationState.media = mediaId;
    NavigationState.search = "";
    NavigationState.page = "home";

    NavigationState.listSplitMode = false;

    updateLayout();

    NavigationState.focus = "left";

    onStatusUpdate();
}

/**
 * Navigates to an album, knowing the media list
 * @param albumId The album ID
 * @param list The media IDs list
 */
export function navigationClickOnAlbumWithList(albumId: number, list: number[]) {
    NavigationState.album = albumId;

    const pos = getCachedAlbumPosition(albumId);

    if (pos < list.length) {
        NavigationState.media = list[pos];
    } else if (list.length > 0) {
        NavigationState.media = list[0];
    } else {
        NavigationState.media = -1;
    }

    NavigationState.page = "home";
    NavigationState.listSplitMode = false;

    updateLayout();

    NavigationState.focus = "left";

    onStatusUpdate(true);
}

/**
 * Changes focus to the left section
 */
export function navigationFocusLeft() {
    NavigationState.focus = "left";

    onStatusUpdate();
}

/**
 * Changes focus to the right section
 */
export function navigationFocusRight() {
    NavigationState.focus = "right";

    onStatusUpdate();
}

/**
 * Closes the page, leaving only the player
 */
export function navigationClosePage() {
    NavigationState.focus = "left";
    NavigationState.listSplitMode = false;
    NavigationState.page = "home";

    updateLayout();

    onStatusUpdate();
}

/**
 * Closes the album, leaving only the player
 */
export function navigationCloseAlbum() {
    NavigationState.focus = "left";
    NavigationState.album = -1;

    updateLayout();
    onStatusUpdate();
}

/**
 * Changes search parameters
 * @param params The new params to set
 */
export function navigationChangeSearchParams(params: string) {
    NavigationState.searchParams = params;

    onStatusUpdate();
}

/**
 * Navigates to the page to find media
 * @param textSearch The text to search
 */
export function navigationGoFindMedia(textSearch: string) {
    tryPreventableExit(() => {
        NavigationState.page = "search";

        NavigationState.album = -1;

        NavigationState.homePageGroup = -1;

        if (NavigationState.media >= 0) {
            NavigationState.listSplitMode = true;
        }

        NavigationState.searchParams = "~" + textSearch;

        updateLayout();

        NavigationState.focus = "right";

        onStatusUpdate();
    });
}

/**
 * Initializes the navigation status
 */
export function initializeNavigation() {
    window.onpopstate = () => {
        loadUrlParams();
        onStatusUpdate();
    };

    loadUrlParams();
    onStatusUpdate();
}
