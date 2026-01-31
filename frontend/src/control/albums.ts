// Albums data controller

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest, makeApiRequest } from "@asanrom/request-browser";
import { setNamedTimeout, clearNamedTimeout } from "@/utils/named-timeouts";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";
import { MediaController } from "./media";
import { setCachedAlbumPosition } from "./player-preferences";
import type { Album, AlbumListItemMin, MediaData, MediaListItem } from "@/api/models";
import { apiAlbumsGetAlbum, apiAlbumsGetAlbumsMin, apiAlbumsMoveMediaInAlbum } from "@/api/api-albums";
import { PagesController } from "./pages";
import { apiMediaGetMedia } from "@/api/api-media";
import {
    addAppEventListener,
    emitAppEvent,
    EVENT_NAME_ALBUMS_CHANGED,
    EVENT_NAME_ALBUMS_LIST_UPDATE,
    EVENT_NAME_APP_STATUS_CHANGED,
    EVENT_NAME_AUTH_CHANGED,
    EVENT_NAME_CURRENT_ALBUM_LOADING,
    EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED,
    EVENT_NAME_CURRENT_ALBUM_UPDATED,
    EVENT_NAME_NEXT_PRE_FETCH,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";

const REQUEST_ID_ALBUMS_LOAD = "albums-load";

const REQUEST_ID_CURRENT_ALBUM_LOAD = "album-current-load";

const REQUEST_ID_NEXT_PRE_FETCH = "album-next-prefetch-load";

/**
 * Minified album data (with lowercased name)
 */
export interface AlbumListItemMinExt {
    /**
     * Album ID
     */
    id: number;

    /**
     * Album name
     */
    name: string;

    /**
     * Album name (lowercase)
     */
    nameLowerCase: string;
}

/**
 * Management object for albums
 */
export class AlbumsController {
    /**
     * Albums mapping
     */
    public static AlbumsMap: Map<number, AlbumListItemMin> = new Map();

    /**
     * Loading flag for albums list
     */
    public static Loading = true;

    /**
     * True if the albums list was loaded at least once
     */
    public static InitiallyLoaded = false;

    /**
     * Initialization logic
     */
    public static Initialize() {
        addAppEventListener(EVENT_NAME_AUTH_CHANGED, AlbumsController.Load);
        addAppEventListener(EVENT_NAME_AUTH_CHANGED, AlbumsController.LoadCurrentAlbum);
        addAppEventListener(EVENT_NAME_APP_STATUS_CHANGED, AlbumsController.OnCurrentAlbumChanged);

        AlbumsController.CurrentAlbum = AppStatus.CurrentAlbum;

        AlbumsController.Load();
        AlbumsController.LoadCurrentAlbum();
    }

    /**
     * Gets the albums list
     * @returns The albums list
     */
    public static GetAlbumsList(): AlbumListItemMin[] {
        return Array.from(AlbumsController.AlbumsMap.values());
    }

    /**
     * Gets a minified version of the albums list
     * @returns A minified version of the albums list
     */
    public static GetAlbumsListMin(): AlbumListItemMinExt[] {
        const result = [];

        for (const album of AlbumsController.AlbumsMap.values()) {
            result.push({
                id: album.id,
                name: album.name,
                nameLowerCase: album.name.toLowerCase(),
            });
        }

        return result;
    }

    /**
     * Finds duplicated name in the map
     * @param name The name to find
     * @returns True if the name os found
     */
    public static FindDuplicatedName(name: string): boolean {
        const nameLower = name.toLowerCase();

        for (const album of AlbumsController.AlbumsMap.values()) {
            if (nameLower === album.name.toLowerCase()) {
                return true;
            }
        }

        return false;
    }

    /**
     * Loads albums list
     */
    public static Load() {
        AlbumsController.Loading = true;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        clearNamedTimeout(REQUEST_ID_ALBUMS_LOAD);
        makeNamedApiRequest(REQUEST_ID_ALBUMS_LOAD, apiAlbumsGetAlbumsMin())
            .onSuccess((albums) => {
                AlbumsController.AlbumsMap.clear();

                for (const album of albums) {
                    AlbumsController.AlbumsMap.set(album.id, album);
                }

                emitAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, AlbumsController.AlbumsMap);
                AlbumsController.Loading = false;
                AlbumsController.InitiallyLoaded = true;
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID_ALBUMS_LOAD, 1500, AlbumsController.Load);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID_ALBUMS_LOAD, 1500, AlbumsController.Load);
            });
    }

    /**
     * Id of the current album.
     * -1 if no album selected.
     */
    public static CurrentAlbum = -1;

    /**
     * Loading flag for the current album
     */
    public static CurrentAlbumLoading = false;

    /**
     * Loaded album data
     */
    public static CurrentAlbumData: Album = null;

    /**
     * Called when the app status changes, in order to reload if necessary
     */
    private static OnCurrentAlbumChanged() {
        if (AppStatus.CurrentAlbum !== AlbumsController.CurrentAlbum) {
            AlbumsController.CurrentAlbum = AppStatus.CurrentAlbum;
            AlbumsController.CurrentAlbumData = null;
            AlbumsController.LoadCurrentAlbum();
        }
        AlbumsController.UpdateAlbumCurrentPos();
    }

    /**
     * Loads the current album
     */
    public static LoadCurrentAlbum() {
        if (AlbumsController.CurrentAlbum < 0) {
            clearNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD);
            abortNamedApiRequest(REQUEST_ID_CURRENT_ALBUM_LOAD);

            AlbumsController.CurrentAlbumData = null;
            emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, null);
            AlbumsController.CurrentAlbumLoading = false;
            emitAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, false);

            AlbumsController.UpdateAlbumCurrentPos();

            return;
        }

        AlbumsController.CurrentAlbumLoading = true;
        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        clearNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD);
        makeNamedApiRequest(REQUEST_ID_CURRENT_ALBUM_LOAD, apiAlbumsGetAlbum(AlbumsController.CurrentAlbum))
            .onSuccess((album) => {
                AlbumsController.CurrentAlbumData = album;
                emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, AlbumsController.CurrentAlbumData);

                AlbumsController.CurrentAlbumLoading = false;
                emitAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, false);

                AlbumsController.UpdateAlbumCurrentPos();
                AppStatus.UpdateURL();
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    notFound: () => {
                        AlbumsController.CurrentAlbumData = null;
                        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, AlbumsController.CurrentAlbumData);

                        AlbumsController.CurrentAlbumLoading = false;
                        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_LOADING, false);

                        AppStatus.CloseAlbum();
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD, 1500, AlbumsController.LoadCurrentAlbum);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID_CURRENT_ALBUM_LOAD, 1500, AlbumsController.LoadCurrentAlbum);
            });
    }

    /**
     * Call when the user makes changes to an album
     * @param albumId The album ID
     * @param noUpdateList Set to true if a list reload is not necessary
     */
    public static OnChangedAlbum(albumId: number, noUpdateList?: boolean) {
        if (AlbumsController.CurrentAlbum === albumId) {
            AlbumsController.LoadCurrentAlbum();
        }
        emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);
        if (!noUpdateList) {
            AlbumsController.Load();
        }
    }

    /**
     * Moves an element of the album to another position.
     * Updates the local list and calls the server to update the remote list.
     * @param oldIndex The original position
     * @param newIndex The new position
     * @param $t Translation function to display errors
     */
    public static MoveCurrentAlbumOrder(oldIndex: number, newIndex: number, $t: (msg: string) => string) {
        if (!AlbumsController.CurrentAlbumData) {
            return;
        }

        if (oldIndex < 0 || oldIndex >= AlbumsController.CurrentAlbumData.list.length) {
            return;
        }

        const albumId = AlbumsController.CurrentAlbumData.id;
        const mediaId = AlbumsController.CurrentAlbumData.list[oldIndex].id;

        AlbumsController.CurrentAlbumData.list.splice(newIndex, 0, AlbumsController.CurrentAlbumData.list.splice(oldIndex, 1)[0]);

        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_UPDATED, AlbumsController.CurrentAlbumData);

        AlbumsController.UpdateAlbumCurrentPos();

        // Update in server

        makeApiRequest(apiAlbumsMoveMediaInAlbum(albumId, mediaId, newIndex))
            .onSuccess(() => {
                emitAppEvent(EVENT_NAME_ALBUMS_CHANGED);
            })
            .onRequestError((err, handleErr) => {
                // Revert changes
                AlbumsController.OnChangedAlbum(albumId, true);
                // Show error
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    maxSizeReached: () => {
                        PagesController.ShowSnackBar(
                            $t("Error") +
                                ": " +
                                $t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                        );
                    },
                    badRequest: () => {
                        PagesController.ShowSnackBar($t("Error") + ": " + $t("Bad request"));
                    },
                    accessDenied: () => {
                        PagesController.ShowSnackBar($t("Error") + ": " + $t("Access denied"));
                        AuthController.CheckAuthStatusSilent();
                    },
                    notFound: () => {
                        PagesController.ShowSnackBar($t("Error") + ": " + $t("Not found"));
                        AlbumsController.OnChangedAlbum(albumId);
                    },
                    serverError: () => {
                        PagesController.ShowSnackBar($t("Error") + ": " + $t("Internal server error"));
                    },
                    networkError: () => {
                        PagesController.ShowSnackBar($t("Error") + ": " + $t("Could not connect to the server"));
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
            });
    }

    /**
     * Position of the current media in the current album
     */
    public static CurrentAlbumPos = -1;

    /**
     * Previous element
     */
    public static CurrentPrev: MediaListItem = null;

    /**
     * Next element
     */
    public static CurrentNext: MediaListItem = null;

    /**
     * Album loop option
     */
    public static AlbumLoop = false;

    /**
     * Album randomize order option
     */
    public static AlbumRandom = false;

    /**
     * Toggles album loop option
     */
    public static ToggleLoop() {
        AlbumsController.AlbumLoop = !AlbumsController.AlbumLoop;
        if (AlbumsController.AlbumLoop) {
            AlbumsController.AlbumRandom = false;
        }
        this.UpdateAlbumCurrentPos();
    }

    /**
     * Toggles album order randomize option
     */
    public static ToggleRandom() {
        AlbumsController.AlbumRandom = !AlbumsController.AlbumRandom;
        if (AlbumsController.AlbumRandom) {
            AlbumsController.AlbumLoop = false;
        }
        this.UpdateAlbumCurrentPos();
    }

    /**
     * Computes and updates the position of the current media in the current album
     */
    public static UpdateAlbumCurrentPos() {
        const mediaId = AppStatus.CurrentMedia;

        if (!AlbumsController.CurrentAlbumData || AlbumsController.CurrentAlbumLoading) {
            AlbumsController.CurrentAlbumPos = -1;
            AlbumsController.CurrentPrev = null;
            AlbumsController.CurrentNext = null;
            emitAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED);
            AlbumsController.PreFetchAlbumNext();
            return;
        }

        if (mediaId < 0 && AlbumsController.CurrentAlbumData.list.length > 0) {
            const albumList = AlbumsController.CurrentAlbumData.list.map((a) => {
                return a.id;
            });
            AppStatus.ClickOnAlbumWithList(AlbumsController.CurrentAlbumData.id, albumList);
            return;
        }

        let mediaPos = -1;

        for (let i = 0; i < AlbumsController.CurrentAlbumData.list.length; i++) {
            if (mediaId === AlbumsController.CurrentAlbumData.list[i].id) {
                mediaPos = i;
                break;
            }
        }

        AlbumsController.CurrentAlbumPos = mediaPos;

        if (mediaPos >= 0) {
            if (AlbumsController.AlbumRandom) {
                let randomIndex = Math.floor(Math.random() * (AlbumsController.CurrentAlbumData.list.length - 1));

                if (randomIndex >= mediaPos) {
                    randomIndex++;
                }

                AlbumsController.CurrentPrev = null;
                AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[randomIndex] || null;

                if (AlbumsController.AlbumLoop) {
                    if (AlbumsController.CurrentNext === null) {
                        AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[0] || null;
                    }
                }
            } else {
                AlbumsController.CurrentPrev = AlbumsController.CurrentAlbumData.list[mediaPos - 1] || null;
                AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[mediaPos + 1] || null;

                if (AlbumsController.AlbumLoop) {
                    if (AlbumsController.CurrentPrev === null) {
                        AlbumsController.CurrentPrev =
                            AlbumsController.CurrentAlbumData.list[AlbumsController.CurrentAlbumData.list.length - 1] || null;
                    }

                    if (AlbumsController.CurrentNext === null) {
                        AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[0] || null;
                    }
                }
            }
        } else {
            AlbumsController.CurrentPrev = null;
            AlbumsController.CurrentNext = null;
        }

        setCachedAlbumPosition(AlbumsController.CurrentAlbumData.id, AlbumsController.CurrentAlbumPos);

        emitAppEvent(EVENT_NAME_CURRENT_ALBUM_MEDIA_POSITION_UPDATED);
        AlbumsController.PreFetchAlbumNext();
    }

    /**
     * True if the next pre-fetch data is available
     */
    public static AvailableNextPrefetch = false;

    /**
     * Loading flag for the next element
     */
    public static LoadingNext = false;

    /**
     * Loaded data for the next element
     */
    public static NextMediaData: MediaData = null;

    /**
     * Loads the data for the next element for fast transition
     */
    public static PreFetchAlbumNext() {
        if (AlbumsController.CurrentNext === null || AlbumsController.CurrentNext.id === MediaController.MediaId) {
            clearNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH);
            abortNamedApiRequest(REQUEST_ID_NEXT_PRE_FETCH);

            AlbumsController.NextMediaData = null;
            AlbumsController.LoadingNext = false;
            AlbumsController.AvailableNextPrefetch = false;
            emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
            return;
        }

        AlbumsController.NextMediaData = null;
        AlbumsController.LoadingNext = true;
        AlbumsController.AvailableNextPrefetch = false;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        const mediaId = AlbumsController.CurrentNext.id;

        clearNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH);
        makeNamedApiRequest(REQUEST_ID_NEXT_PRE_FETCH, apiMediaGetMedia(mediaId))
            .onSuccess((media) => {
                AlbumsController.NextMediaData = media;
                AlbumsController.LoadingNext = false;
                AlbumsController.AvailableNextPrefetch = true;
                emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
            })
            .onRequestError((err, handleErr) => {
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    notFound: () => {
                        AlbumsController.NextMediaData = null;
                        AlbumsController.LoadingNext = false;
                        AlbumsController.AvailableNextPrefetch = true;
                        emitAppEvent(EVENT_NAME_NEXT_PRE_FETCH);
                    },
                    temporalError: () => {
                        // Retry
                        setNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH, 1500, AlbumsController.PreFetchAlbumNext);
                    },
                });
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                setNamedTimeout(REQUEST_ID_NEXT_PRE_FETCH, 1500, AlbumsController.PreFetchAlbumNext);
            });
    }

    /**
     * Checks if pre-fetch is available
     * @returns True if the pre-fetched data was set, so no load is needed
     */
    public static CheckAlbumNextPrefetch(): boolean {
        if (AlbumsController.CurrentNext === null) {
            return false;
        }

        if (AlbumsController.CurrentNext.id !== MediaController.MediaId) {
            return false;
        }

        if (AlbumsController.AvailableNextPrefetch) {
            MediaController.SetMediaData(AlbumsController.NextMediaData);
            return true;
        } else {
            return false;
        }
    }
}
