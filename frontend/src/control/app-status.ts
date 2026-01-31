// App status
// Controls what page the app is in
// Event: app-status-update

"use strict";

import { getParameterByName } from "@/utils/cookie";
import { AlbumsController } from "./albums";
import { emitAppEvent, EVENT_NAME_APP_STATUS_CHANGED, EVENT_NAME_MEDIA_DELETE } from "./app-events";
import { getCachedAlbumPosition } from "./player-preferences";
import { generateURIQuery } from "@/utils/api";
import { ExitPreventer } from "./exit-prevent";

/**
 * Layout mode
 */
export type AppStatusLayout = "initial" | "album" | "media" | "media-split";

/**
 * Focus target
 */
export type AppStatusFocus = "right" | "left";

/**
 * Page
 */
export type AppStatusPage = "home" | "media" | "random" | "random" | "albums" | "upload" | "search";

/**
 * App status manager object
 */
export class AppStatus {
    /**
     * Current layout mode
     */
    public static CurrentLayout: AppStatusLayout = "initial";

    /**
     * Current focused side
     */
    public static CurrentFocus: AppStatusFocus = "left";

    /**
     * Current main page
     */
    public static CurrentPage: AppStatusPage = "home";

    /**
     * Current search query
     */
    public static CurrentSearch = "";

    /**
     * Current search params
     */
    public static SearchParams = "";

    /**
     * Random seed
     */
    public static RandomSeed = 0;

    /**
     * True for split mode, false for single mode
     */
    public static ListSplitMode = true;

    /**
     * Current media ID
     */
    public static CurrentMedia = -1;

    /**
     * Current album ID
     */
    public static CurrentAlbum = -1;

    /**
     * Current home page group ID
     */
    public static CurrentHomePageGroup = -1;

    /**
     * Initialization logic
     * Loads at the app startup
     */
    public static Initialize() {
        window.onpopstate = function () {
            AppStatus.LoadURLParams();
            AppStatus.OnStatusUpdate();
        };

        AppStatus.LoadURLParams();
        AppStatus.OnStatusUpdate();
    }

    /**
     * Loads parameters from URL
     */
    public static LoadURLParams() {
        const media = getParameterByName("media");
        if (media) {
            const mediaId = parseInt(media);

            if (!isNaN(mediaId) && mediaId >= 0) {
                AppStatus.CurrentMedia = mediaId;
            } else {
                AppStatus.CurrentMedia = -1;
            }
        } else {
            AppStatus.CurrentMedia = -1;
        }

        const album = getParameterByName("album");
        if (album) {
            const albumId = parseInt(album);

            if (!isNaN(albumId) && albumId >= 0) {
                AppStatus.CurrentAlbum = albumId;
            } else {
                AppStatus.CurrentAlbum = -1;
            }
        } else {
            AppStatus.CurrentAlbum = -1;
        }

        const page = getParameterByName("page");

        if (page && ["home", "media", "random", "albums", "upload", "search"].includes(page)) {
            AppStatus.CurrentPage = page as AppStatusPage;
        } else {
            AppStatus.CurrentPage = "home";
        }

        const search = getParameterByName("search");

        if (search) {
            AppStatus.CurrentSearch = search;
        } else {
            AppStatus.CurrentSearch = "";
        }

        AppStatus.CurrentHomePageGroup = -1;

        if (AppStatus.CurrentPage === "random") {
            AppStatus.RandomSeed = Math.floor(parseInt(getParameterByName("seed") || "0", 10)) || Date.now();
        } else if (AppStatus.CurrentPage === "home" && AppStatus.CurrentMedia >= 0) {
            const group = getParameterByName("g");
            if (group) {
                const groupId = parseInt(group);

                if (!isNaN(groupId) && groupId >= 0) {
                    AppStatus.CurrentHomePageGroup = groupId;
                }
            }
        }

        const searchParams = getParameterByName("sp");

        if (searchParams) {
            AppStatus.SearchParams = searchParams;
        } else {
            AppStatus.SearchParams = "";
        }

        const split = getParameterByName("split");

        AppStatus.ListSplitMode = split === "yes";
    }

    /**
     * Updates layout mode based on current status variables,
     * like: page, media, album, split mode
     */
    public static UpdateLayout() {
        if (AppStatus.CurrentMedia >= 0) {
            if (AppStatus.CurrentAlbum >= 0) {
                // Media with album list
                AppStatus.CurrentLayout = "album";
            } else if (AppStatus.ListSplitMode) {
                // Media with list
                AppStatus.CurrentLayout = "media-split";
            } else {
                // Media alone
                AppStatus.CurrentLayout = "media";
            }
        } else if (AppStatus.CurrentAlbum >= 0) {
            AppStatus.CurrentLayout = "album";
        } else {
            // Just initial, sidebar + List
            AppStatus.CurrentLayout = "initial";
        }
    }

    /**
     * Checks if player is visible
     * @returns True if player is visible
     */
    public static IsPlayerVisible(): boolean {
        switch (AppStatus.CurrentLayout) {
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
    public static IsPageVisible(): boolean {
        switch (AppStatus.CurrentLayout) {
            case "initial":
            case "media-split":
                return true;
            default:
                return false;
        }
    }

    /**
     * Generates an URL from the current status
     * @returns The current URL
     */
    public static GetCurrentURL(): string {
        const params: any = Object.create(null);

        if (AppStatus.CurrentMedia >= 0) {
            params["media"] = AppStatus.CurrentMedia + "";
        }

        if (AppStatus.CurrentAlbum >= 0) {
            params["album"] = AppStatus.CurrentAlbum + "";
        }

        if (AppStatus.CurrentPage && AppStatus.CurrentPage !== "home") {
            params["page"] = AppStatus.CurrentPage;
        }

        if (AppStatus.CurrentPage === "home") {
            if (AppStatus.CurrentHomePageGroup >= 0 && AppStatus.CurrentMedia >= 0) {
                params["g"] = AppStatus.CurrentHomePageGroup + "";
            }
        }

        if (AppStatus.CurrentSearch) {
            params["search"] = AppStatus.CurrentSearch;
        }

        if (AppStatus.SearchParams) {
            params["sp"] = AppStatus.SearchParams;
        }

        if (AppStatus.RandomSeed && AppStatus.CurrentPage === "random") {
            params["seed"] = AppStatus.RandomSeed + "";
        }

        if (AppStatus.ListSplitMode) {
            params["split"] = "yes";
        }

        return window.location.protocol + "//" + window.location.host + window.location.pathname + generateURIQuery(params);
    }

    /**
     * Updates current URL
     * @param replaceState True if the new URL must replace the old URL
     */
    public static UpdateURL(replaceState?: boolean) {
        if (AppStatus.CurrentAlbum >= 0 && AppStatus.CurrentMedia < 0) {
            if (AlbumsController.Loading) {
                return;
            }
            if (AlbumsController.CurrentAlbum !== AppStatus.CurrentAlbum) {
                return;
            }
            if (AlbumsController.CurrentAlbumData && AlbumsController.CurrentAlbumData.list.length > 0) {
                return;
            }
        }

        if (history.pushState) {
            const newURL = AppStatus.GetCurrentURL();
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
    private static OnStatusUpdate(replaceState?: boolean) {
        AppStatus.UpdateLayout();

        emitAppEvent(EVENT_NAME_APP_STATUS_CHANGED);

        AppStatus.UpdateURL(replaceState);
    }

    /**
     * Navigates to a page
     * @param page The page to navigate to
     */
    public static GoToPage(page: AppStatusPage) {
        ExitPreventer.TryExit(() => {
            AppStatus.CurrentPage = page;

            if (AppStatus.CurrentPage === "random") {
                AppStatus.RandomSeed = Date.now();
            }

            AppStatus.CurrentAlbum = -1;

            AppStatus.CurrentHomePageGroup = -1;

            if (AppStatus.CurrentMedia >= 0) {
                AppStatus.ListSplitMode = true;
            }

            AppStatus.SearchParams = "";

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";

            AppStatus.OnStatusUpdate();
        });
    }

    /**
     * Navigates to a page
     * Sets split mode only if there is a media selected, and no split enabled
     * @param page The page to navigate to
     * @param searchParams Search parameters
     */
    public static GoToPageConditionalSplit(page: AppStatusPage, searchParams?: string) {
        ExitPreventer.TryExit(() => {
            const changedPage = AppStatus.CurrentPage !== page;

            AppStatus.CurrentPage = page;

            if (AppStatus.CurrentPage === "random") {
                AppStatus.RandomSeed = Date.now();
            }

            if (AppStatus.CurrentMedia !== -1 && AppStatus.CurrentAlbum === -1 && (!AppStatus.ListSplitMode || changedPage)) {
                AppStatus.ListSplitMode = true;
            } else {
                AppStatus.CurrentMedia = -1;
                AppStatus.ListSplitMode = false;
            }

            AppStatus.CurrentAlbum = -1;

            AppStatus.SearchParams = searchParams || "";

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";

            AppStatus.OnStatusUpdate();
        });
    }

    /**
     * Expands the page, closing the player
     */
    public static ExpandPage() {
        ExitPreventer.TryExit(() => {
            AppStatus.CurrentAlbum = -1;
            AppStatus.CurrentMedia = -1;
            AppStatus.CurrentHomePageGroup = -1;
            AppStatus.ListSplitMode = false;

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";

            AppStatus.OnStatusUpdate();
        });
    }

    /**
     * Call when the current media is deleted
     */
    public static OnDeleteMedia() {
        AppStatus.CurrentMedia = -1;
        AppStatus.CurrentHomePageGroup = -1;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        emitAppEvent(EVENT_NAME_MEDIA_DELETE);

        AppStatus.OnStatusUpdate();
    }

    /**
     * Generates a random new seed
     */
    public static RefreshSeed() {
        AppStatus.RandomSeed = Date.now();
        AppStatus.OnStatusUpdate();
    }

    /**
     * Changes the search query
     * @param search The search query
     * @param forced True to force the page change to the search results
     */
    public static GoToSearch(search: string, forced?: boolean) {
        AppStatus.CurrentSearch = search;

        if (AppStatus.CurrentPage === "random") {
            AppStatus.RandomSeed = Date.now();
        }

        if (AppStatus.CurrentSearch) {
            if (forced || AppStatus.CurrentPage !== "random") {
                AppStatus.CurrentPage = "media";
            }

            if (AppStatus.CurrentMedia >= 0) {
                AppStatus.ListSplitMode = true;
            }

            AppStatus.CurrentAlbum = -1;

            AppStatus.SearchParams = "";

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";
        }

        AppStatus.OnStatusUpdate();
    }

    /**
     * Clears current search query
     */
    public static ClearSearch() {
        AppStatus.CurrentSearch = "";

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    /**
     * Navigates to a media asset
     * @param mediaId The media ID
     * @param split True to use split mode
     * @param group ID of the home group
     */
    public static ClickOnMedia(mediaId: number, split: boolean, group?: number) {
        ExitPreventer.TryExit(() => {
            AppStatus.CurrentMedia = mediaId;

            if (split) {
                AppStatus.ListSplitMode = true;

                if (typeof group === "number" && group >= 0) {
                    AppStatus.CurrentHomePageGroup = group;
                }
            }

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "left";

            AppStatus.OnStatusUpdate();
        });
    }

    /**
     * Navigates to an album
     * @param albumId The album ID
     */
    public static ClickOnAlbum(albumId: number) {
        ExitPreventer.TryExit(() => {
            AppStatus.CurrentAlbum = albumId;
            AppStatus.CurrentMedia = -1;
            AppStatus.CurrentHomePageGroup = -1;
            AppStatus.CurrentPage = "home";

            AppStatus.ListSplitMode = false;

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "left";

            AppStatus.OnStatusUpdate();
        });
    }

    /**
     * Navigates to an album, while keeping the current media
     * @param albumId The album ID
     * @param mediaId The media ID
     */
    public static ClickOnAlbumByMedia(albumId: number, mediaId: number) {
        AppStatus.CurrentAlbum = albumId;
        AppStatus.CurrentMedia = mediaId;
        AppStatus.CurrentSearch = "";
        AppStatus.CurrentPage = "home";

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    /**
     * navigates to an album, knowing the media list
     * @param albumId The album ID
     * @param list The media IDs list
     */
    public static ClickOnAlbumWithList(albumId: number, list: number[]) {
        AppStatus.CurrentAlbum = albumId;

        const pos = getCachedAlbumPosition(albumId);

        if (pos < list.length) {
            AppStatus.CurrentMedia = list[pos];
        } else if (list.length > 0) {
            AppStatus.CurrentMedia = list[0];
        } else {
            AppStatus.CurrentMedia = -1;
        }

        AppStatus.CurrentPage = "home";
        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate(true);
    }

    /**
     * Changes focus to the left section
     */
    public static FocusLeft() {
        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    /**
     * Changes focus to the right section
     */
    public static FocusRight() {
        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    /**
     * Closes the page, leaving only the player
     */
    public static ClosePage() {
        AppStatus.CurrentFocus = "left";
        AppStatus.ListSplitMode = false;
        AppStatus.CurrentPage = "home";
        AppStatus.UpdateLayout();
        AppStatus.OnStatusUpdate();
    }

    /**
     * Closes the album, leaving only the player
     */
    public static CloseAlbum() {
        AppStatus.CurrentFocus = "left";
        AppStatus.CurrentAlbum = -1;
        AppStatus.UpdateLayout();
        AppStatus.OnStatusUpdate();
    }

    /**
     * Changes search parameters
     * @param params The new params to set
     */
    public static ChangeSearchParams(params: string) {
        AppStatus.SearchParams = params;

        AppStatus.OnStatusUpdate();
    }

    /**
     * Navigates to the page to find media
     * @param textSearch The text to search
     */
    public static GoFindMedia(textSearch: string) {
        ExitPreventer.TryExit(() => {
            AppStatus.CurrentPage = "search";

            AppStatus.CurrentAlbum = -1;

            AppStatus.CurrentHomePageGroup = -1;

            if (AppStatus.CurrentMedia >= 0) {
                AppStatus.ListSplitMode = true;
            }

            AppStatus.SearchParams = "~" + textSearch;

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";

            AppStatus.OnStatusUpdate();
        });
    }
}
